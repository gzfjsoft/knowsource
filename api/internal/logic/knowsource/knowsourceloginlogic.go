package knowsource

import (
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"knowsource/api/internal/superadmin"
	"knowsource/api/internal/svc"
	"knowsource/api/internal/types"
	"knowsource/api/internal/utils"
	"knowsource/common/cryptx"
	"knowsource/common/jwtx"
	"knowsource/common/response"
	"knowsource/consts"
	"knowsource/model"

	"github.com/zeromicro/go-zero/core/logx"
)

type KnowsourceLoginLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 员工登录
func NewKnowsourceLoginLogic(ctx context.Context, svcCtx *svc.ServiceContext) *KnowsourceLoginLogic {
	return &KnowsourceLoginLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *KnowsourceLoginLogic) KnowsourceLogin(req *types.KnowsourceLoginRequest) (resp *types.KnowsourceLoginResponse, err error) {
	// 验证参数
	if strings.TrimSpace(req.ClientId) == "" {
		return &types.KnowsourceLoginResponse{
			Response: types.Response{
				Code:    response.ParameterErrorCode,
				Message: "clientId不能为空",
			},
		}, nil
	}
	clientId := strings.TrimSpace(req.ClientId)

	// 校验 client 是否存在
	clientRow, cErr := l.svcCtx.ClientModel.FindOneByClientId(l.ctx, clientId)
	if cErr != nil {
		if cErr == model.ErrNotFound {
			return &types.KnowsourceLoginResponse{
				Response: types.Response{
					Code:    response.NotFoundCode,
					Message: "client 不存在",
				},
			}, nil
		}
		return &types.KnowsourceLoginResponse{
			Response: types.Response{
				Code:    response.ServerErrorCode,
				Message: "查询 client 失败",
				Info:    cErr.Error(),
			},
		}, nil
	}

	if clientRow.Status == 0 {
		return &types.KnowsourceLoginResponse{
			Response: types.Response{
				Code:    response.EmailNotVerifiedCode,
				Message: "租户尚未完成邮箱验证，请先使用「验证邮箱」接口完成验证后再登录",
			},
		}, nil
	}
	if clientRow.Status == 2 {
		return &types.KnowsourceLoginResponse{
			Response: types.Response{
				Code:    response.ForbiddenCode,
				Message: "租户已停用",
			},
		}, nil
	}

	// 自动确保该 clientId 存在 superadmin（缺则创建）
	if _, se := superadmin.EnsureSuperadmin(l.ctx, l.svcCtx, clientId); se != nil {
		logx.Errorf("EnsureSuperadmin failed clientId=%s err=%v", clientId, se)
	}

	if req.EmpCode == "" || req.Password == "" {
		return &types.KnowsourceLoginResponse{
			Response: types.Response{
				Code:    response.ParameterErrorCode,
				Message: "员工编码和密码不能为空",
			},
		}, nil
	}

	// 验证验证码（必填）
	if !utils.VerifyCaptcha(req.CaptchaId, req.Captcha) {
		logx.Infof("验证码错误: %s, %s", req.CaptchaId, req.Captcha)
		return &types.KnowsourceLoginResponse{
			Response: types.Response{
				Code:    response.WrongCaptchaCode,
				Message: "验证码错误",
			},
		}, nil
	}

	// 查找员工
	emp, err := l.svcCtx.FrEmpModel.FindOneByClientIdFempCode(l.ctx, clientId, req.EmpCode)
	if err != nil {
		if err == model.ErrNotFound {
			return &types.KnowsourceLoginResponse{
				Response: types.Response{
					Code:    response.UserNotExistCode,
					Message: "员工不存在",
				},
			}, nil
		}
		return &types.KnowsourceLoginResponse{
			Response: types.Response{
				Code:    response.ServerErrorCode,
				Message: "查询员工失败",
				Info:    "FrEmpModel.FindOne " + err.Error(),
			},
		}, nil
	}

	// 检查员工状态
	if emp.Status != 0 {
		return &types.KnowsourceLoginResponse{
			Response: types.Response{
				Code:    response.ForbiddenCode,
				Message: "员工账号未启用或已注销",
			},
		}, nil
	}

	// 查找密码（按 tenant）
	empPassword, err := l.svcCtx.EmpPasswordModel.FindOneByClientIdEmpCode(l.ctx, clientId, req.EmpCode)

	if err == model.ErrNotFound {
		return &types.KnowsourceLoginResponse{
			Response: types.Response{
				Code:    response.UserNotExistCode,
				Message: "员工密码未设置",
				Info:    clientId + "+" + req.EmpCode,
			},
		}, nil
	} else if err != nil {
		return &types.KnowsourceLoginResponse{
			Response: types.Response{
				Code:    response.ServerErrorCode,
				Message: "查询密码失败",
				Info:    "EmpPasswordModel.FindOneByClientIdEmpCode " + err.Error(),
			},
		}, nil
	}

	// 验证密码
	passwordHash := cryptx.PasswordEncrypt(l.svcCtx.Config.Salt, req.Password)
	if empPassword.Password != passwordHash {
		return &types.KnowsourceLoginResponse{
			Response: types.Response{
				Code:    response.ForbiddenCode,
				Message: "密码错误",
			},
		}, nil
	}

	// 查找部门信息
	var deptName string
	if emp.DeptCode != "" {
		dept, err := l.svcCtx.FrDeptModel.FindOneByClientIdDeptCode(l.ctx, clientId, emp.DeptCode)
		if err == nil {
			deptName = dept.DeptName
		}
	}

	// 生成 JWT token
	// 使用 GenerateTokenWithContext 生成 token，确保 role 字段是 int64 类型
	// userId: 0 (knowsource 系统暂时不使用 userId)
	// empId: 员工编码
	// isAdmin: 0 (普通用户，0:普通用户 1:普通管理员 2:超级管理员)
	// companyId: 0 (暂时不使用)
	// userName: 员工姓名
	// orgId: 0 (暂时不使用)
	// roleIds: "" (暂时不使用)

	accessExpire := l.svcCtx.Config.Auth.AccessExpire
	if req.IsDebug == 1 {
		accessExpire = 3600 * 24 * 30 // 30天
	}

	expireDuration := time.Duration(accessExpire) * time.Second

	frRols, err := l.svcCtx.FrUserRolesModel.FindAllByClientIdEmpCode(l.ctx, clientId, emp.FempCode)
	if err != nil {
		return &types.KnowsourceLoginResponse{
			Response: types.Response{
				Code:    response.ServerErrorCode,
				Message: "查询用户角色失败",
				Info:    "FrUserRolesModel.FindAllByClientIdEmpCode " + err.Error(),
			},
		}, nil
	}

	// 如果用户没有角色，自动插入 "user" 角色
	if len(frRols) == 0 {
		_, err = l.svcCtx.FrUserRolesModel.Insert(l.ctx, &model.FrUserRoles{
			ClientId: clientId,
			EmpCode:  emp.FempCode,
			Role:     "user",
		})
		if err != nil {
			return &types.KnowsourceLoginResponse{
				Response: types.Response{
					Code:    response.ServerErrorCode,
					Message: "创建用户角色失败",
					Info:    "FrUserRolesModel.Insert " + err.Error(),
				},
			}, nil
		}
		// 重新查询角色列表
		frRols, err = l.svcCtx.FrUserRolesModel.FindAllByClientIdEmpCode(l.ctx, clientId, emp.FempCode)
		if err != nil {
			return &types.KnowsourceLoginResponse{
				Response: types.Response{
					Code:    response.ServerErrorCode,
					Message: "查询用户角色失败",
					Info:    "FrUserRolesModel.FindAllByClientIdEmpCode " + err.Error(),
				},
			}, nil
		}
	}

	roleSlice, roles := frUserRolesToCodes(frRols)
	isAdmin := 0
	for _, code := range roleSlice {
		if code == consts.ONLY_ADMIN || code == consts.SUPER_ADMIN {
			isAdmin = 1
			break
		}
	}

	empPermissions := []string{}
	frPermissions, err := l.svcCtx.FrRolesPermissionsModel.FindAllByClientIdEmpCode(l.ctx, clientId, emp.FempCode)
	if err != nil {
		return &types.KnowsourceLoginResponse{
			Response: types.Response{
				Code:    response.ServerErrorCode,
				Message: "查询用户权限失败",
				Info:    "FrRolesPermissionsModel.FindAllByClientIdEmpCode " + err.Error(),
			},
		}, nil
	}
	for _, permission := range frPermissions {
		empPermissions = append(empPermissions, permission.Permission)
	}

	// 将权限列表写入 Redis
	permissionsKey := fmt.Sprintf("user:permissions:%s:%s", clientId, emp.FempCode)
	permissionsJSON, err := json.Marshal(empPermissions)
	if err != nil {
		logx.Errorf("序列化权限列表失败: %v, empCode: %s", err, emp.FempCode)
		return &types.KnowsourceLoginResponse{
			Response: types.Response{
				Code:    response.ServerErrorCode,
				Message: "序列化权限列表失败",
				Info:    err.Error(),
			},
		}, nil
	}
	// 设置过期时间为 token 过期时间的 2 倍，确保权限列表不会在 token 有效期内过期
	expireSeconds := int(expireDuration.Seconds() * 2)
	err = l.svcCtx.RedisClient.Setex(permissionsKey, string(permissionsJSON), expireSeconds)
	if err != nil {
		logx.Errorf("写入权限列表到Redis失败: %v, empCode: %s", err, emp.FempCode)
		// 不阻止登录，只记录错误
	}

	// 将 clientId 写入 context，供 jwtx/model GenToken 等读取
	ctxWithClient := context.WithValue(l.ctx, "clientId", clientId)
	token, err := jwtx.GenerateTokenWithContext(ctxWithClient, clientId, 0, emp.FempCode, int64(isAdmin), emp.FempName, roles, expireDuration)
	if err != nil {
		return &types.KnowsourceLoginResponse{
			Response: types.Response{
				Code:    response.ServerErrorCode,
				Message: "生成token失败",
				Info:    "jwtx.GenerateTokenWithContext " + err.Error(),
			},
		}, nil
	}

	// 构建用户信息
	clientName := ""
	clientDesp := ""
	if clientRow != nil && strings.TrimSpace(clientRow.ClientJsonInfo) != "" {
		var meta struct {
			Name string `json:"name"`
			Desp string `json:"desp"`
		}
		_ = json.Unmarshal([]byte(clientRow.ClientJsonInfo), &meta)
		clientName = strings.TrimSpace(meta.Name)
		clientDesp = strings.TrimSpace(meta.Desp)
	}
	if clientName == "" {
		clientName = filepath.Base(clientId)
	}
	userInfo := types.KnowsourceEmpInfo{
		EmpCode:        emp.FempCode,
		EmpName:        emp.FempName,
		ClientId:       clientId,
		ClientName:     clientName,
		ClientDesp:     clientDesp,
		DeptCode:       emp.DeptCode,
		DeptName:       deptName,
		Status:         emp.Status,
		Position:       emp.Fposition,
		EmpPermissions: empPermissions,
		Roles:          roleSlice,
	}
	if m := strings.TrimSpace(emp.Mobile); m != "" {
		userInfo.Mobile = m
	}
	if e := strings.TrimSpace(emp.Email); e != "" {
		userInfo.Email = e
	}

	return &types.KnowsourceLoginResponse{
		Response: types.Response{
			Code:    response.SuccessCode,
			Message: "登录成功",
		},
		Data: &types.KnowsourceLoginData{
			Token:    token,
			UserInfo: userInfo,
		},
	}, nil
}

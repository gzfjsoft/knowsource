package knowsource

import (
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"knowsource/api/internal/svc"
	"knowsource/api/internal/types"
	"knowsource/common/jwtx"
	"knowsource/common/response"
	"knowsource/consts"
	"knowsource/model"

	"github.com/zeromicro/go-zero/core/logx"
)

type RefreshTokenLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewRefreshTokenLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RefreshTokenLogic {
	return &RefreshTokenLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *RefreshTokenLogic) RefreshToken() (resp *types.KnowsourceLoginResponse, err error) {
	// 从 context 中获取员工编码
	empCodeValue := l.ctx.Value("empCode")
	if empCodeValue == nil {
		return &types.KnowsourceLoginResponse{
			Response: types.Response{
				Code:    response.UnauthorizedCode,
				Message: "未登录或登录已过期",
			},
		}, nil
	}

	empCode, ok := empCodeValue.(string)
	if !ok || empCode == "" {
		return &types.KnowsourceLoginResponse{
			Response: types.Response{
				Code:    response.UnauthorizedCode,
				Message: "员工编码无效",
			},
		}, nil
	}

	clientId, _ := l.ctx.Value("clientId").(string)
	if strings.TrimSpace(clientId) == "" {
		return &types.KnowsourceLoginResponse{
			Response: types.Response{
				Code:    response.UnauthorizedCode,
				Message: "clientId不能为空，请重新登录",
			},
		}, nil
	}

	// 查找员工
	emp, err := l.svcCtx.FrEmpModel.FindOneByClientIdFempCode(l.ctx, clientId, empCode)
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

	// 查找密码信息（用于获取角色）
	// empPassword, err := l.svcCtx.EmpPasswordModel.FindOneByClientIdEmpCode(l.ctx, clientId, empCode)
	// if err != nil {
	// 	if err == model.ErrNotFound {
	// 		return &types.KnowsourceLoginResponse{
	// 			Response: types.Response{
	// 				Code:    response.UserNotExistCode,
	// 				Message: "员工密码未设置",
	// 			},
	// 		}, nil
	// 	}
	// 	return &types.KnowsourceLoginResponse{
	// 		Response: types.Response{
	// 			Code:    response.ServerErrorCode,
	// 			Message: "查询密码失败",
	// 			Info:    "EmpPasswordModel.FindOneByClientIdEmpCode " + err.Error(),
	// 		},
	// 	}, nil
	// }

	// 查找部门信息
	var deptName string
	if emp.DeptCode != "" {
		dept, err := l.svcCtx.FrDeptModel.FindOneByClientIdDeptCode(l.ctx, clientId, emp.DeptCode)
		if err == nil {
			deptName = dept.DeptName
		}
	}

	clientName := ""
	clientDesp := ""
	clientRow, cErr := l.svcCtx.ClientModel.FindOneByClientId(l.ctx, clientId)
	if cErr == nil && clientRow != nil && strings.TrimSpace(clientRow.ClientJsonInfo) != "" {
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
	expireDuration = time.Duration(l.svcCtx.Config.Auth.AccessExpire) * time.Second
	expireSeconds := int(expireDuration.Seconds() * 2)
	err = l.svcCtx.RedisClient.Setex(permissionsKey, string(permissionsJSON), expireSeconds)
	if err != nil {
		logx.Errorf("写入权限列表到Redis失败: %v, empCode: %s", err, emp.FempCode)
		// 不阻止刷新token，只记录错误
	}

	token, err := jwtx.GenerateTokenWithContext(l.ctx, clientId, 0, emp.FempCode, int64(isAdmin), emp.FempName, roles, expireDuration)
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
			Message: "刷新token成功",
		},
		Data: &types.KnowsourceLoginData{
			Token:    token,
			UserInfo: userInfo,
		},
	}, nil
}

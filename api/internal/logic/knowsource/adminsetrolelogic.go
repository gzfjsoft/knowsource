package knowsource

import (
	"context"
	"strings"

	"knowsource/api/internal/svc"
	"knowsource/api/internal/types"
	"knowsource/api/internal/utils"
	"knowsource/common/response"
	"knowsource/consts"
	"knowsource/model"

	"github.com/zeromicro/go-zero/core/logx"
)

type AdminSetRoleLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// admin 设置角色
func NewAdminSetRoleLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AdminSetRoleLogic {
	return &AdminSetRoleLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *AdminSetRoleLogic) AdminSetRole(req *types.AdminSetRoleRequest) (resp *types.Response, err error) {
	// 验证参数
	if req.EmpCode == "" {
		return &types.Response{
			Code:    response.ParameterErrorCode,
			Message: "员工编码不能为空",
		}, nil
	}

	if req.Role == "" {
		return &types.Response{
			Code:    response.ParameterErrorCode,
			Message: "角色不能为空",
		}, nil
	}

	// 验证角色值
	if req.Role != consts.ONLY_ADMIN && req.Role != "user" && req.Role != consts.SUPER_ADMIN {
		return &types.Response{
			Code:    response.ParameterErrorCode,
			Message: "角色值无效，只能是 superadmin, admin 或 user",
		}, nil
	}

	clientId, _ := l.ctx.Value("clientId").(string)
	if strings.TrimSpace(clientId) == "" {
		return &types.Response{
			Code:    response.UnauthorizedCode,
			Message: "clientId不能为空，请重新登录",
		}, nil
	}

	// 检查当前用户是否是 admin
	if !utils.IsAdminRoleFromContext(l.ctx) {
		role, _ := utils.GetRoleFromContext(l.ctx)
		return &types.Response{
			Code:    response.UnauthorizedCode,
			Message: "只有管理员才能设置角色",
			Info:    role,
		}, nil
	}

	// 检查员工是否存在
	_, err = l.svcCtx.FrEmpModel.FindOneByClientIdFempCode(l.ctx, clientId, req.EmpCode)
	if err != nil {
		if err == model.ErrNotFound {
			return &types.Response{
				Code:    response.UserNotExistCode,
				Message: "员工不存在",
			}, nil
		}
		l.Errorf("查询员工失败: %v", err)
		return &types.Response{
			Code:    response.ServerErrorCode,
			Message: "查询员工失败: " + err.Error(),
		}, nil
	}

	// 查找员工密码记录（按 tenant）
	empPassword, err := l.svcCtx.EmpPasswordModel.FindOneByClientIdEmpCode(l.ctx, clientId, req.EmpCode)
	if err != nil {
		if err == model.ErrNotFound {
			// 如果密码记录不存在，创建新记录（使用默认密码或空密码）
			empPassword = &model.EmpPassword{
				ClientId: clientId,
				EmpCode:  req.EmpCode,
				Password: "", // 密码为空，需要用户首次登录时设置

			}
			_, err = l.svcCtx.EmpPasswordModel.Insert(l.ctx, empPassword)
			if err != nil {
				l.Errorf("创建密码记录失败: %v", err)
				return &types.Response{
					Code:    response.ServerErrorCode,
					Message: "创建密码记录失败: " + err.Error(),
				}, nil
			}
		} else {
			l.Errorf("查询员工密码失败: %v", err)
			return &types.Response{
				Code:    response.ServerErrorCode,
				Message: "查询员工密码失败: " + err.Error(),
			}, nil
		}
	} else {
		// 如果密码记录存在，更新角色

		err = l.svcCtx.EmpPasswordModel.Update(l.ctx, empPassword)
		if err != nil {
			l.Errorf("更新角色失败: %v", err)
			return &types.Response{
				Code:    response.ServerErrorCode,
				Message: "更新角色失败: " + err.Error(),
			}, nil
		}
	}

	// 获取当前管理员信息用于日志
	adminEmpCode := l.ctx.Value("empCode")
	adminEmpCodeStr := ""
	if adminEmpCode != nil {
		adminEmpCodeStr, _ = adminEmpCode.(string)
	}

	l.Infof("管理员 %s 将员工 %s 的角色设置为 %s", adminEmpCodeStr, req.EmpCode, req.Role)

	return &types.Response{
		Code:    response.SuccessCode,
		Message: "角色设置成功",
	}, nil
}

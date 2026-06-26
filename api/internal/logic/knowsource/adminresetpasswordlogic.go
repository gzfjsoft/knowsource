package knowsource

import (
	"context"
	"strings"
	"unicode"

	"knowsource/api/internal/svc"
	"knowsource/api/internal/types"
	"knowsource/api/internal/utils"
	"knowsource/common/cryptx"
	"knowsource/common/response"
	"knowsource/model"

	"github.com/zeromicro/go-zero/core/logx"
)

type AdminResetPasswordLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// admin 重置密码
func NewAdminResetPasswordLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AdminResetPasswordLogic {
	return &AdminResetPasswordLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *AdminResetPasswordLogic) AdminResetPassword(req *types.AdminResetPasswordRequest) (resp *types.Response, err error) {
	// 验证参数
	if req.EmpCode == "" {
		return &types.Response{
			Code:    response.ParameterErrorCode,
			Message: "员工编码不能为空",
		}, nil
	}

	if req.Password == "" {
		return &types.Response{
			Code:    response.ParameterErrorCode,
			Message: "新密码不能为空",
		}, nil
	}

	// 验证密码强度：至少8位，包含大小写字母、数字和符号
	if !validatePasswordStrength(req.Password) {
		return &types.Response{
			Code:    response.ParameterErrorCode,
			Message: "密码格式不正确，应至少8位，包含大小写字母、数字和特殊符号",
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
			Message: "只有管理员或超级管理员才能重置密码",
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
			// 如果密码记录不存在，创建新记录
			newPasswordHash := cryptx.PasswordEncrypt(l.svcCtx.Config.Salt, req.Password)
			empPassword = &model.EmpPassword{
				ClientId: clientId,
				EmpCode:  req.EmpCode,
				Password: newPasswordHash,
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
		// 如果密码记录存在，更新密码
		newPasswordHash := cryptx.PasswordEncrypt(l.svcCtx.Config.Salt, req.Password)
		empPassword.Password = newPasswordHash
		err = l.svcCtx.EmpPasswordModel.Update(l.ctx, empPassword)
		if err != nil {
			l.Errorf("更新密码失败: %v", err)
			return &types.Response{
				Code:    response.ServerErrorCode,
				Message: "更新密码失败: " + err.Error(),
			}, nil
		}
	}

	// 获取当前管理员信息用于日志
	adminEmpCode := l.ctx.Value("empCode")
	adminEmpCodeStr := ""
	if adminEmpCode != nil {
		adminEmpCodeStr, _ = adminEmpCode.(string)
	}

	l.Infof("管理员 %s 重置了员工 %s 的密码", adminEmpCodeStr, req.EmpCode)

	return &types.Response{
		Code:    response.SuccessCode,
		Message: "密码重置成功",
	}, nil
}

// validatePasswordStrength 验证密码强度
// 要求：至少8位，包含大小写字母、数字和特殊符号
func validatePasswordStrength(password string) bool {
	// 检查长度：至少8位
	if len(password) < 8 {
		return false
	}

	hasUpper := false   // 大写字母
	hasLower := false   // 小写字母
	hasDigit := false   // 数字
	hasSpecial := false // 特殊符号

	for _, c := range password {
		switch {
		case unicode.IsUpper(c):
			hasUpper = true
		case unicode.IsLower(c):
			hasLower = true
		case unicode.IsDigit(c):
			hasDigit = true
		case unicode.IsPunct(c) || unicode.IsSymbol(c):
			hasSpecial = true
		}
	}

	// 必须同时包含大小写字母、数字和特殊符号
	return hasUpper && hasLower && hasDigit && hasSpecial
}

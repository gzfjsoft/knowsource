package logic

import (
	"context"

	"knowsource/api/internal/svc"
	"knowsource/api/internal/types"
	"knowsource/api/internal/utils"
	"knowsource/common/response"
	"knowsource/consts"

	"github.com/zeromicro/go-zero/core/logx"
)

type AdminUserUpdateLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewAdminUserUpdateLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AdminUserUpdateLogic {
	return &AdminUserUpdateLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *AdminUserUpdateLogic) AdminUserUpdate(req *types.AdminUserUpdateRequest) (resp *types.Response, err error) {
	// todo: add your logic here and delete this line
	// Initialize response
	resp = &types.Response{
		Code:    response.SuccessCode,
		Message: "success",
	}

	if !utils.IsAdminRoleFromContext(l.ctx) {
		role, _ := utils.GetRoleFromContext(l.ctx)
		resp.Code = response.UnauthorizedCode
		resp.Message = "只有超级管理员和系统管理员才有权限修改用户信息."
		resp.Info = role
		return resp, nil
	}

	if req.SysRole == consts.SUPER_ADMIN && !utils.IsSuperAdminRoleFromContext(l.ctx) {
		role, _ := utils.GetRoleFromContext(l.ctx)
		resp.Code = response.UnauthorizedCode
		resp.Message = "只有超级管理员才有权限修改用户信息为超级管理员."
		resp.Info = role
		return resp, nil
	}

	if req.SysRole != "" && (req.SysRole != consts.SUPER_ADMIN && req.SysRole != consts.ONLY_ADMIN && req.SysRole != "user" && req.SysRole != "member") {
		resp.Code = response.ParameterErrorCode
		resp.Message = "invalid sysRole, only superadmin, admin, user, member are allowed"
		resp.Info = req.SysRole
		return resp, nil
	}

	// Validate user ID exists
	if req.UserId == 0 {
		resp.Code = response.ParameterErrorCode
		resp.Message = "找不到用户"
		resp.Info = "userId = 0"
		return resp, nil
	}

	// Get existing user
	user, err := l.svcCtx.UsersModel.FindOne(l.ctx, req.UserId)
	if err != nil {
		resp.Code = response.UserNotExistCode
		resp.Message = "找不到用户"
		resp.Info = err.Error()
		return resp, err
	}

	if req.SysRole != "" {
		user.SysRole = req.SysRole
	}

	// Update fields if provided
	if req.Username != "" {
		user.Username = req.Username
	}
	if req.Nickname != "" {
		user.Nickname = req.Nickname
	}
	if req.Phone != "" {
		user.Phone = req.Phone
	}
	if req.Email != "" {
		user.Email = req.Email
	}

	if req.IsDeleted != 0 {
		user.IsDeleted = req.IsDeleted
	}

	// Save updates
	err = l.svcCtx.UsersModel.Update(l.ctx, user)
	if err != nil {
		resp.Code = 500
		resp.Message = "更新用户失败"
		resp.Info = err.Error()
		return resp, err
	}

	return
}

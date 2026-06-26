package logic

import (
	"context"
	"knowsource/api/internal/svc"
	"knowsource/api/internal/types"
	"knowsource/common/response"
	"knowsource/consts"
	"knowsource/model"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetUserRolesLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetUserRolesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetUserRolesLogic {
	return &GetUserRolesLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetUserRolesLogic) GetUserRoles(userId uint64) (resp *types.UserRolesResponse) {

	sysrole, _ := l.ctx.Value("role").(string)
	if sysrole != consts.ONLY_ADMIN && sysrole != consts.SUPER_ADMIN {
		return &types.UserRolesResponse{
			Response: types.Response{
				Code:    response.UnauthorizedCode,
				Message: "没有权限",
			},
		}
	}

	resp = &types.UserRolesResponse{
		Response: types.Response{},
	}

	// Check if user exists
	_, err := l.svcCtx.UsersModel.FindOne(l.ctx, userId)
	if err != nil {
		if err == model.ErrNotFound {
			resp.Response = types.Response{
				Code:    response.InvalidRequestParamCode,
				Message: "User not found",
			}
			return resp
		}
		resp.Response = types.Response{
			Code:    response.ServerErrorCode,
			Message: err.Error(),
		}
		return resp
	}

	// Get user roles
	userRoles, err := l.svcCtx.UserRolesModel.FindByUserId(l.ctx, userId)
	if err != nil {
		resp.Response = types.Response{
			Code:    response.ServerErrorCode,
			Message: err.Error(),
		}
		return resp
	}

	var roles []types.Role
	for _, userRole := range userRoles {
		role, err := l.svcCtx.RolesModel.FindOne(l.ctx, uint64(userRole.RoleId))
		if err != nil {
			continue // Skip if role not found
		}
		roles = append(roles, types.Role{
			Id:          role.RoleId,
			Name:        role.RoleName,
			Description: role.Description.String,
			CreatedAt:   role.CreatedAt.Unix(),
			UpdatedAt:   role.UpdatedAt.Unix(),
		})
	}

	resp.Response = types.Response{
		Code:    response.SuccessCode,
		Message: "Success",
	}
	resp.Data = types.UserRolesResponseData{
		Roles: roles,
	}

	return resp
}

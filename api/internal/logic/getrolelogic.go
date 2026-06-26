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

type GetRoleLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetRoleLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetRoleLogic {
	return &GetRoleLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetRoleLogic) GetRole(id uint64) (resp *types.GetRoleResponse) {

	sysrole, _ := l.ctx.Value("role").(string)
	if sysrole != consts.ONLY_ADMIN && sysrole != consts.SUPER_ADMIN {
		return &types.GetRoleResponse{
			Response: types.Response{
				Code:    response.UnauthorizedCode,
				Message: "没有权限",
			},
		}
	}
	resp = &types.GetRoleResponse{
		Response: types.Response{},
	}
	role, err := l.svcCtx.RolesModel.FindOne(l.ctx, id)
	if err != nil {
		if err == model.ErrNotFound {
			resp.Response = types.Response{
				Code:    response.InvalidRequestParamCode,
				Message: "Role not found",
			}
			return resp
		}
		resp.Response = types.Response{
			Code:    response.ServerErrorCode,
			Message: err.Error(),
		}
		return resp
	}

	// Get permissions for this role
	rolePermissions, err := l.svcCtx.RolePermissionsModel.FindByRoleId(l.ctx, id)
	if err != nil {
		resp.Response = types.Response{
			Code:    response.ServerErrorCode,
			Message: err.Error(),
		}
		return resp
	}

	var permissions []types.Permission
	for _, rp := range rolePermissions {
		permission, err := l.svcCtx.PermissionsModel.FindOneByPermissionName(l.ctx, rp.PermissionName)
		if err != nil {
			continue // Skip if permission not found
		}
		permissions = append(permissions, types.Permission{
			Id:          permission.PermissionId,
			Name:        permission.PermissionName,
			Description: permission.Description.String,
			CreatedAt:   permission.CreatedAt.Unix(),
		})
	}

	resp.Response = types.Response{
		Code:    response.SuccessCode,
		Message: "Success",
	}
	resp.Data = types.GetRoleResponseData{
		Role: types.Role{
			Id:          role.RoleId,
			Name:        role.RoleName,
			Description: role.Description.String,
			CreatedAt:   role.CreatedAt.Unix(),
			UpdatedAt:   role.UpdatedAt.Unix(),
		},
		Permissions: permissions,
	}

	return resp
}

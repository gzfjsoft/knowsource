package logic

import (
	"context"
	"database/sql"
	"knowsource/api/internal/svc"
	"knowsource/api/internal/types"
	"knowsource/common/response"
	"knowsource/consts"
	"knowsource/model"
	"time"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type UpdateRoleLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUpdateRoleLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateRoleLogic {
	return &UpdateRoleLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UpdateRoleLogic) UpdateRole(req *types.UpdateRoleRequest) (resp *types.RoleResponse) {

	sysrole, _ := l.ctx.Value("role").(string)
	if sysrole != consts.ONLY_ADMIN && sysrole != consts.SUPER_ADMIN {
		return &types.RoleResponse{
			Response: types.Response{
				Code:    response.UnauthorizedCode,
				Message: "only admin can access",
			},
		}
	}

	resp = &types.RoleResponse{
		Response: types.Response{},
	}

	// Check if role exists
	role, err := l.svcCtx.RolesModel.FindOne(l.ctx, req.Id)
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

	// If updating name, check if new name already exists
	if req.Name != "" && req.Name != role.RoleName {
		existingRole, err := l.svcCtx.RolesModel.FindOneByRoleName(l.ctx, req.Name)
		if err != nil && err != model.ErrNotFound {
			resp.Response = types.Response{
				Code:    response.ServerErrorCode,
				Message: err.Error(),
			}
			return resp
		}
		if existingRole != nil {
			resp.Response = types.Response{
				Code:    response.InvalidRequestParamCode,
				Message: "Role name already exists",
			}
			return resp
		}
		role.RoleName = req.Name
	}

	if req.Description != "" {
		role.Description = sql.NullString{
			String: req.Description,
			Valid:  true,
		}
	}

	role.UpdatedAt = time.Now()

	err = sqlx.NewMysql(l.svcCtx.Config.MySQL.DataSource).Transact(func(session sqlx.Session) error {
		// Update role with transaction
		rolesModel := l.svcCtx.RolesModel.WithSession(session)
		err := rolesModel.Update(l.ctx, role)
		if err != nil {
			return err
		}

		// Update permissions if provided
		if len(req.Permissions) > 0 {
			rolePermissionsModel := l.svcCtx.RolePermissionsModel.WithSession(session)

			// Delete existing permissions
			err = rolePermissionsModel.DeleteByRoleId(l.ctx, role.RoleId)
			if err != nil {
				return err
			}

			// Insert new permissions
			for _, permissionName := range req.Permissions {
				_, err = rolePermissionsModel.Insert(l.ctx, &model.RolePermissions{
					RoleId:         uint64(role.RoleId),
					PermissionName: permissionName,
					GrantedAt:      time.Now(),
				})
				if err != nil {
					return err
				}
			}
		}

		return nil
	})

	if err != nil {
		resp.Response = types.Response{
			Code:    response.ServerErrorCode,
			Message: err.Error(),
		}
		return resp
	}

	resp.Response = types.Response{
		Code:    response.SuccessCode,
		Message: "Success",
	}
	resp.Data = types.Role{
		Id:          role.RoleId,
		Name:        role.RoleName,
		Description: role.Description.String,
		CreatedAt:   role.CreatedAt.Unix(),
		UpdatedAt:   role.UpdatedAt.Unix(),
	}

	return resp
}

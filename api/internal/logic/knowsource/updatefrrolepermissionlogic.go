package knowsource

import (
	"context"
	"strings"

	"knowsource/api/internal/svc"
	"knowsource/api/internal/types"
	"knowsource/common/response"
	"knowsource/model"

	"github.com/pkg/errors"
	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateFrRolePermissionLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 更新角色权限关联
func NewUpdateFrRolePermissionLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateFrRolePermissionLogic {
	return &UpdateFrRolePermissionLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UpdateFrRolePermissionLogic) UpdateFrRolePermission(req *types.FrRolePermissionUpdateRequest) (resp *types.Response, err error) {
	if req.Id == 0 {
		return &types.Response{
			Code:    response.ParameterErrorCode,
			Message: "关联ID不能为空",
		}, nil
	}

	if req.Role == "" {
		return &types.Response{
			Code:    response.ParameterErrorCode,
			Message: "角色编码不能为空",
		}, nil
	}

	if req.Permission == "" {
		return &types.Response{
			Code:    response.ParameterErrorCode,
			Message: "权限编码不能为空",
		}, nil
	}

	clientId, _ := l.ctx.Value("clientId").(string)
	if strings.TrimSpace(clientId) == "" {
		return &types.Response{
			Code:    response.UnauthorizedCode,
			Message: "clientId不能为空，请重新登录",
		}, nil
	}

	// 检查关联是否存在
	_, err = l.svcCtx.FrRolesPermissionsModel.FindOne(l.ctx, req.Id)
	if err != nil {
		if errors.Is(err, model.ErrNotFound) {
			return &types.Response{
				Code:    response.NotFoundCode,
				Message: "角色权限关联不存在",
			}, nil
		}
		l.Logger.Errorf("查询角色权限关联失败: %v", err)
		return &types.Response{
			Code:    response.ServerErrorCode,
			Message: "Database error",
			Info:    err.Error(),
		}, nil
	}

	// 检查是否已存在相同的角色和权限组合（排除当前记录）
	existing, err := l.svcCtx.FrRolesPermissionsModel.FindOneByClientIdRolePermission(l.ctx, clientId, req.Role, req.Permission)
	if err == nil && existing.Id != req.Id {
		return &types.Response{
			Code:    response.ParameterErrorCode,
			Message: "该角色和权限组合已存在",
		}, nil
	}

	if err != nil && !errors.Is(err, model.ErrNotFound) {
		l.Logger.Errorf("查询角色权限关联失败: %v", err)
		return &types.Response{
			Code:    response.ServerErrorCode,
			Message: "Database error",
			Info:    err.Error(),
		}, nil
	}

	// 更新关联
	rolePermission := &model.FrRolesPermissions{
		ClientId:   clientId,
		Id:         req.Id,
		Role:       req.Role,
		Permission: req.Permission,
	}

	err = l.svcCtx.FrRolesPermissionsModel.Update(l.ctx, rolePermission)
	if err != nil {
		l.Logger.Errorf("更新角色权限关联失败: %v", err)
		return &types.Response{
			Code:    response.ServerErrorCode,
			Message: "Database error",
			Info:    err.Error(),
		}, nil
	}

	return &types.Response{
		Code:    response.SuccessCode,
		Message: "success",
	}, nil
}

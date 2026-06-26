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

type CreateFrRolePermissionLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 创建角色权限关联
func NewCreateFrRolePermissionLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateFrRolePermissionLogic {
	return &CreateFrRolePermissionLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CreateFrRolePermissionLogic) CreateFrRolePermission(req *types.FrRolePermissionCreateRequest) (resp *types.Response, err error) {
	clientId, _ := l.ctx.Value("clientId").(string)
	if strings.TrimSpace(clientId) == "" {
		return &types.Response{
			Code:    response.UnauthorizedCode,
			Message: "clientId不能为空，请重新登录",
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

	// 检查是否已存在相同的角色和权限组合
	_, err = l.svcCtx.FrRolesPermissionsModel.FindOneByClientIdRolePermission(l.ctx, clientId, req.Role, req.Permission)
	if err == nil {
		return &types.Response{
			Code:    response.ParameterErrorCode,
			Message: "该角色和权限组合已存在",
		}, nil
	}

	if !errors.Is(err, model.ErrNotFound) {
		l.Logger.Errorf("查询角色权限关联失败: %v", err)
		return &types.Response{
			Code:    response.ServerErrorCode,
			Message: "Database error",
			Info:    err.Error(),
		}, nil
	}

	// 创建新关联
	rolePermission := &model.FrRolesPermissions{
		ClientId:   clientId,
		Role:       req.Role,
		Permission: req.Permission,
	}

	_, err = l.svcCtx.FrRolesPermissionsModel.Insert(l.ctx, rolePermission)
	if err != nil {
		l.Logger.Errorf("创建角色权限关联失败: %v", err)
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

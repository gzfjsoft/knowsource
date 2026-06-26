package knowsource

import (
	"context"

	"knowsource/api/internal/svc"
	"knowsource/api/internal/types"
	"knowsource/common/response"
	"knowsource/model"

	"github.com/pkg/errors"
	"github.com/zeromicro/go-zero/core/logx"
)

type GetFrRolePermissionLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 获取角色权限关联
func NewGetFrRolePermissionLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetFrRolePermissionLogic {
	return &GetFrRolePermissionLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetFrRolePermissionLogic) GetFrRolePermission(req *types.FrRolePermissionGetRequest) (resp *types.FrRolePermissionGetResponse, err error) {
	if req.Id == 0 {
		return &types.FrRolePermissionGetResponse{
			Response: types.Response{
				Code:    response.ParameterErrorCode,
				Message: "关联ID不能为空",
			},
		}, nil
	}

	rolePermission, err := l.svcCtx.FrRolesPermissionsModel.FindOne(l.ctx, req.Id)
	if err != nil {
		if errors.Is(err, model.ErrNotFound) {
			return &types.FrRolePermissionGetResponse{
				Response: types.Response{
					Code:    response.NotFoundCode,
					Message: "角色权限关联不存在",
				},
			}, nil
		}
		l.Logger.Errorf("查询角色权限关联失败: %v", err)
		return &types.FrRolePermissionGetResponse{
			Response: types.Response{
				Code:    response.ServerErrorCode,
				Message: "Database error",
				Info:    err.Error(),
			},
		}, nil
	}

	return &types.FrRolePermissionGetResponse{
		Response: types.Response{
			Code:    response.SuccessCode,
			Message: "success",
		},
		Data: &types.FrRolePermissionInfo{
			Id:         rolePermission.Id,
			Role:       rolePermission.Role,
			Permission: rolePermission.Permission,
		},
	}, nil
}

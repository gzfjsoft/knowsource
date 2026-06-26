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

type DeleteFrRolePermissionLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 删除角色权限关联
func NewDeleteFrRolePermissionLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteFrRolePermissionLogic {
	return &DeleteFrRolePermissionLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *DeleteFrRolePermissionLogic) DeleteFrRolePermission(req *types.FrRolePermissionDeleteRequest) (resp *types.Response, err error) {
	if req.Id == 0 {
		return &types.Response{
			Code:    response.ParameterErrorCode,
			Message: "关联ID不能为空",
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

	// 删除关联
	err = l.svcCtx.FrRolesPermissionsModel.Delete(l.ctx, req.Id)
	if err != nil {
		l.Logger.Errorf("删除角色权限关联失败: %v", err)
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

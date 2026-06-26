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

type DeleteFrUserRoleLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 删除用户角色关联
func NewDeleteFrUserRoleLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteFrUserRoleLogic {
	return &DeleteFrUserRoleLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *DeleteFrUserRoleLogic) DeleteFrUserRole(req *types.FrUserRoleDeleteRequest) (resp *types.Response, err error) {
	if req.Id == 0 {
		return &types.Response{
			Code:    response.ParameterErrorCode,
			Message: "关联ID不能为空",
		}, nil
	}

	// 检查关联是否存在
	_, err = l.svcCtx.FrUserRolesModel.FindOne(l.ctx, req.Id)
	if err != nil {
		if errors.Is(err, model.ErrNotFound) {
			return &types.Response{
				Code:    response.NotFoundCode,
				Message: "用户角色关联不存在",
			}, nil
		}
		l.Logger.Errorf("查询用户角色关联失败: %v", err)
		return &types.Response{
			Code:    response.ServerErrorCode,
			Message: "Database error",
			Info:    err.Error(),
		}, nil
	}

	// 删除关联
	err = l.svcCtx.FrUserRolesModel.Delete(l.ctx, req.Id)
	if err != nil {
		l.Logger.Errorf("删除用户角色关联失败: %v", err)
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

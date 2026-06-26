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

type UpdateFrPermissionLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 更新权限
func NewUpdateFrPermissionLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateFrPermissionLogic {
	return &UpdateFrPermissionLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UpdateFrPermissionLogic) UpdateFrPermission(req *types.FrPermissionUpdateRequest) (resp *types.Response, err error) {
	clientId, _ := l.ctx.Value("clientId").(string)
	clientId = strings.TrimSpace(clientId)
	if clientId == "" {
		return &types.Response{Code: response.UnauthorizedCode, Message: "clientId不能为空，请重新登录"}, nil
	}

	if req.Permission == "" {
		return &types.Response{
			Code:    response.ParameterErrorCode,
			Message: "权限编码不能为空",
		}, nil
	}

	// 检查权限是否存在
	existing, err := l.svcCtx.FrPermissionsModel.FindOneByPermission(l.ctx, req.Permission)
	if err != nil {
		if errors.Is(err, model.ErrNotFound) {
			return &types.Response{
				Code:    response.NotFoundCode,
				Message: "权限不存在",
			}, nil
		}
		l.Logger.Errorf("查询权限失败: %v", err)
		return &types.Response{
			Code:    response.ServerErrorCode,
			Message: "Database error",
			Info:    err.Error(),
		}, nil
	}

	// 更新权限
	permission := &model.FrPermissions{
		Id:          existing.Id,
		Permission:  req.Permission,
		Description: req.Description,
	}

	err = l.svcCtx.FrPermissionsModel.Update(l.ctx, permission)
	if err != nil {
		l.Logger.Errorf("更新权限失败: %v", err)
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

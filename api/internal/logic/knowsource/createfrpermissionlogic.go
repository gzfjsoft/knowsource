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

type CreateFrPermissionLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 创建权限
func NewCreateFrPermissionLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateFrPermissionLogic {
	return &CreateFrPermissionLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CreateFrPermissionLogic) CreateFrPermission(req *types.FrPermissionCreateRequest) (resp *types.Response, err error) {
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

	// 检查权限是否已存在
	_, err = l.svcCtx.FrPermissionsModel.FindOneByPermission(l.ctx, req.Permission)
	if err == nil {
		return &types.Response{
			Code:    response.ParameterErrorCode,
			Message: "权限编码已存在",
		}, nil
	}

	if !errors.Is(err, model.ErrNotFound) {
		l.Logger.Errorf("查询权限失败: %v", err)
		return &types.Response{
			Code:    response.ServerErrorCode,
			Message: "Database error",
			Info:    err.Error(),
		}, nil
	}

	// 创建新权限
	permission := &model.FrPermissions{
		Permission:  req.Permission,
		Description: req.Description,
	}

	_, err = l.svcCtx.FrPermissionsModel.Insert(l.ctx, permission)
	if err != nil {
		l.Logger.Errorf("创建权限失败: %v", err)
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

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

type GetFrPermissionLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 获取权限
func NewGetFrPermissionLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetFrPermissionLogic {
	return &GetFrPermissionLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetFrPermissionLogic) GetFrPermission(req *types.FrPermissionGetRequest) (resp *types.FrPermissionGetResponse, err error) {
	clientId, _ := l.ctx.Value("clientId").(string)
	clientId = strings.TrimSpace(clientId)
	if clientId == "" {
		return &types.FrPermissionGetResponse{
			Response: types.Response{Code: response.UnauthorizedCode, Message: "clientId不能为空，请重新登录"},
		}, nil
	}

	if req.Permission == "" {
		return &types.FrPermissionGetResponse{
			Response: types.Response{
				Code:    response.ParameterErrorCode,
				Message: "权限编码不能为空",
			},
		}, nil
	}

	permission, err := l.svcCtx.FrPermissionsModel.FindOneByPermission(l.ctx, req.Permission)
	if err != nil {
		if errors.Is(err, model.ErrNotFound) {
			return &types.FrPermissionGetResponse{
				Response: types.Response{
					Code:    response.NotFoundCode,
					Message: "权限不存在",
				},
			}, nil
		}
		l.Logger.Errorf("查询权限失败: %v", err)
		return &types.FrPermissionGetResponse{
			Response: types.Response{
				Code:    response.ServerErrorCode,
				Message: "Database error",
				Info:    err.Error(),
			},
		}, nil
	}

	return &types.FrPermissionGetResponse{
		Response: types.Response{
			Code:    response.SuccessCode,
			Message: "success",
		},
		Data: &types.FrPermissionInfo{
			Permission:  permission.Permission,
			Description: permission.Description,
		},
	}, nil
}

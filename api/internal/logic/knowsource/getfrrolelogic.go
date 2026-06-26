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

type GetFrRoleLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 获取角色
func NewGetFrRoleLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetFrRoleLogic {
	return &GetFrRoleLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetFrRoleLogic) GetFrRole(req *types.FrRoleGetRequest) (resp *types.FrRoleGetResponse, err error) {
	clientId, _ := l.ctx.Value("clientId").(string)
	clientId = strings.TrimSpace(clientId)
	if clientId == "" {
		return &types.FrRoleGetResponse{
			Response: types.Response{
				Code:    response.UnauthorizedCode,
				Message: "clientId不能为空，请重新登录",
			},
		}, nil
	}

	if req.Role == "" {
		return &types.FrRoleGetResponse{
			Response: types.Response{
				Code:    response.ParameterErrorCode,
				Message: "角色编码不能为空",
			},
		}, nil
	}

	role, err := l.svcCtx.FrRolesModel.FindOneByClientIdRole(l.ctx, clientId, req.Role)
	if err != nil {
		if errors.Is(err, model.ErrNotFound) {
			return &types.FrRoleGetResponse{
				Response: types.Response{
					Code:    response.NotFoundCode,
					Message: "角色不存在",
				},
			}, nil
		}
		l.Logger.Errorf("查询角色失败: %v", err)
		return &types.FrRoleGetResponse{
			Response: types.Response{
				Code:    response.ServerErrorCode,
				Message: "Database error",
				Info:    err.Error(),
			},
		}, nil
	}

	return &types.FrRoleGetResponse{
		Response: types.Response{
			Code:    response.SuccessCode,
			Message: "success",
		},
		Data: &types.FrRoleInfo{
			Role: role.Role,
			Name: role.Name,
		},
	}, nil
}

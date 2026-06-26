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

type GetDifyOptionLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// get dify option
func NewGetDifyOptionLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetDifyOptionLogic {
	return &GetDifyOptionLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetDifyOptionLogic) GetDifyOption(req *types.GetDifyOptionReq) (resp *types.GetDifyOptionResponse, err error) {
	clientId, _ := l.ctx.Value("clientId").(string)
	clientId = strings.TrimSpace(clientId)
	if clientId == "" {
		return &types.GetDifyOptionResponse{
			Response: types.Response{Code: response.UnauthorizedCode, Message: "clientId不能为空，请重新登录"},
		}, nil
	}

	if req.Name == "" {
		return &types.GetDifyOptionResponse{
			Response: types.Response{
				Code:    response.ParameterErrorCode,
				Message: "name参数不能为空",
			},
		}, nil
	}

	option, err := l.svcCtx.DifyOptionsModel.FindOneByClientIdName(l.ctx, clientId, req.Name)
	if err != nil {
		if errors.Is(err, model.ErrNotFound) {
			return &types.GetDifyOptionResponse{
				Response: types.Response{
					Code:    response.NotFoundCode,
					Message: "记录不存在",
				},
			}, nil
		}
		return &types.GetDifyOptionResponse{
			Response: types.Response{
				Code:    response.ServerErrorCode,
				Message: "Database error",
				Info:    err.Error(),
			},
		}, nil
	}

	return &types.GetDifyOptionResponse{
		Response: types.Response{
			Code:    response.SuccessCode,
			Message: "success",
		},
		Data: &types.GetDifyOptionData{
			Name:        option.Name,
			Url:         option.Url,
			ApiKey:      option.ApiKey,
			Description: option.Description,
		},
	}, nil
}

package knowsource

import (
	"context"
	"strings"

	"knowsource/api/internal/svc"
	"knowsource/api/internal/types"
	"knowsource/common/response"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListDifyOptionLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// List dify option
func NewListDifyOptionLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListDifyOptionLogic {
	return &ListDifyOptionLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ListDifyOptionLogic) ListDifyOption() (resp *types.GetDifyOptionListResponse, err error) {
	clientId, _ := l.ctx.Value("clientId").(string)
	clientId = strings.TrimSpace(clientId)
	if clientId == "" {
		return &types.GetDifyOptionListResponse{
			Response: types.Response{
				Code:    response.UnauthorizedCode,
				Message: "clientId不能为空，请重新登录",
			},
		}, nil
	}

	options, err := l.svcCtx.DifyOptionsModel.FindAllByClientId(l.ctx, clientId)
	if err != nil {
		return &types.GetDifyOptionListResponse{
			Response: types.Response{
				Code:    response.ServerErrorCode,
				Message: "Database error",
				Info:    err.Error(),
			},
		}, nil
	}

	list := make([]types.GetDifyOptionData, len(options))
	for i, option := range options {

		list[i] = types.GetDifyOptionData{
			Name:        option.Name,
			Url:         option.Url,
			ApiKey:      option.ApiKey,
			Description: option.Description,
		}
	}

	return &types.GetDifyOptionListResponse{
		Response: types.Response{
			Code:    response.SuccessCode,
			Message: "success",
		},
		Data: &types.GetDifyOptionListData{
			List:  list,
			Total: int64(len(list)),
		},
	}, nil
}

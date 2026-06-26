package ai

import (
	"context"

	"knowsource/api/internal/svc"
	"knowsource/api/internal/types"

	"knowsource/common/response"

	"github.com/zeromicro/go-zero/core/logx"
)

type AISessionChatLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 会话聊天
func NewAISessionChatLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AISessionChatLogic {
	return &AISessionChatLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *AISessionChatLogic) AISessionChat(req *types.AISessionChatRequest) (resp *types.AISessionChatResponse, err error) {
	return &types.AISessionChatResponse{
		Response: types.Response{
			Code:    response.ParameterErrorCode,
			Message: "没有实现这个功能",
		},
		Data: types.AISessionChatData{
			Response: "",
		},
	}, nil
}

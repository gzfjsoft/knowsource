package ai

import (
	"context"

	"knowsource/api/internal/svc"
	"knowsource/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type AISessionChatOptionsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 会话聊天选项
func NewAISessionChatOptionsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AISessionChatOptionsLogic {
	return &AISessionChatOptionsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *AISessionChatOptionsLogic) AISessionChatOptions(req *types.AISessionChatRequest) (resp *types.AISessionChatResponse, err error) {
	// todo: add your logic here and delete this line

	return
}

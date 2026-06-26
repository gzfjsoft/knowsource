package ai

import (
	"context"

	"knowsource/api/internal/svc"
	"knowsource/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type AIChatLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// AI聊天
func NewAIChatLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AIChatLogic {
	return &AIChatLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *AIChatLogic) AIChat(req *types.OllamaRequest) (resp *types.OllamaResponse, err error) {
	// todo: add your logic here and delete this line

	return
}

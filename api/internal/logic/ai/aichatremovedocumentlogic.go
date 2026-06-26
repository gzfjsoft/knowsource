package ai

import (
	"context"
	"strings"

	"knowsource/api/internal/svc"
	"knowsource/api/internal/types"
	"knowsource/api/internal/utils"
	"knowsource/common/response"

	"github.com/zeromicro/go-zero/core/logx"
)

type AIChatRemoveDocumentLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// AI对话删除已上传的临时文档（按文件名从缓存移除）
func NewAIChatRemoveDocumentLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AIChatRemoveDocumentLogic {
	return &AIChatRemoveDocumentLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *AIChatRemoveDocumentLogic) AIChatRemoveDocument(req *types.AIChatRemoveDocumentRequest) (resp *types.Response, err error) {
	empCode, _ := l.ctx.Value("empCode").(string)
	if empCode == "" {
		return &types.Response{
			Code:    response.UnauthorizedCode,
			Message: "未登录",
		}, nil
	}
	filename := strings.TrimSpace(req.Filename)
	if filename == "" {
		return &types.Response{
			Code:    response.InvalidRequestParamCodeInHandler,
			Message: "文件名不能为空",
		}, nil
	}
	if l.svcCtx.RedisClient != nil {
		_ = utils.AIChatDocCacheRemove(l.svcCtx.RedisClient, empCode, filename)
	}
	return &types.Response{
		Code:    200,
		Message: "success",
	}, nil
}

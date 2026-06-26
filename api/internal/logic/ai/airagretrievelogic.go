// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package ai

import (
	"context"
	"strings"

	"knowsource/api/internal/svc"
	"knowsource/api/internal/types"
	"knowsource/common/response"

	"github.com/zeromicro/go-zero/core/logx"
)

type AIRagRetrieveLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// RAG 检索预览（不走聊天模型）
func NewAIRagRetrieveLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AIRagRetrieveLogic {
	return &AIRagRetrieveLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *AIRagRetrieveLogic) AIRagRetrieve(req *types.AIRagRetrieveRequest) (resp *types.AIRagRetrieveResponse, err error) {
	if strings.TrimSpace(req.Message) == "" {
		return &types.AIRagRetrieveResponse{
			Response: types.Response{
				Code:    response.ServerErrorCode,
				Message: "message 不能为空",
			},
		}, nil
	}
	if strings.TrimSpace(req.DocumentCode) == "" {
		return &types.AIRagRetrieveResponse{
			Response: types.Response{
				Code:    response.ServerErrorCode,
				Message: "documentCode 不能为空",
			},
		}, nil
	}

	clientId, _ := l.ctx.Value("clientId").(string)
	clientId = strings.TrimSpace(clientId)

	orchestrator := NewRagRetrievalOrchestrator(l.ctx, l.svcCtx)
	result, retrieveErr := orchestrator.Retrieve(RagRetrieveRequest{
		ClientId:        clientId,
		Message:         req.Message,
		DocumentCode:    req.DocumentCode,
		Tags:            req.Tags,
		SkipRag:         req.Skiprag,
		HasUploadedDocs: false,
	})
	if retrieveErr != nil {
		l.Errorf("RAG 检索预览失败: %v", retrieveErr)
		return &types.AIRagRetrieveResponse{
			Response: types.Response{
				Code:    response.ServerErrorCode,
				Message: "RAG 检索失败",
				Info:    retrieveErr.Error(),
			},
		}, nil
	}

	data := &types.AIRagRetrieveData{}
	if result != nil {
		docs := make([]types.AIRagDocumentItem, 0, len(result.Documents))
		for _, d := range result.Documents {
			docs = append(docs, types.AIRagDocumentItem{
				Source:          d.Source,
				Content:         d.Content,
				FileName:        d.FileName,
				Path:            d.Path,
				RawDocId:        d.RawDocId,
				Page:            d.Page,
				QdrantId:        d.QdrantId,
				SimilarityScore: d.SimilarityScore,
				RerankScore:     d.RerankScore,
			})
		}
		data = &types.AIRagRetrieveData{
			Documents:        docs,
			RawSearchUsed:    result.RawSearchUsed,
			VectorSearchUsed: result.VectorSearchUsed,
			FullTextMs:       result.FullTextMs,
			MainSearchMs:     result.MainSearchMs,
			SubSearchMs:      result.SubSearchMs,
			RerankConfigured: result.RerankConfigured,
			SummaryRerank:    airagRerankDiagFrom(result.SummaryRerank),
			ChunkRerank:      airagRerankDiagFrom(result.ChunkRerank),
		}
	}

	return &types.AIRagRetrieveResponse{
		Response: types.Response{
			Code:    response.SuccessCode,
			Message: "success",
		},
		Data: data,
	}, nil
}

func airagRerankDiagFrom(src *types.SearchRawDocVectorsRerankInfo) *types.AIRagRerankDiag {
	if src == nil {
		return nil
	}
	return &types.AIRagRerankDiag{
		UsedRerank:    src.UsedRerank,
		Error:         src.Error,
		RerankTopK:    src.RerankTopK,
		OriginalCount: src.OriginalCount,
		RerankedCount: src.RerankedCount,
	}
}

// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package knowsource

import (
	"context"
	"net/http"
	"strings"
	"time"

	"knowsource/api/internal/svc"
	"knowsource/api/internal/types"
	"knowsource/api/internal/utils"
	"knowsource/common/response"

	"github.com/zeromicro/go-zero/core/logx"
)

type TestLLMServicesLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 测试 Embedding 与 Rerank 实际调用（租户覆盖配置）
func NewTestLLMServicesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *TestLLMServicesLogic {
	return &TestLLMServicesLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *TestLLMServicesLogic) TestLLMServices(req *types.LLMServiceTestRequest) (resp *types.LLMServiceTestResponse, err error) {
	clientId, _ := l.ctx.Value("clientId").(string)
	clientId = strings.TrimSpace(clientId)
	if clientId == "" {
		return &types.LLMServiceTestResponse{
			Response: types.Response{
				Code:    response.ParameterErrorCode,
				Message: "clientId 不能为空",
			},
		}, nil
	}

	query := strings.TrimSpace(req.Query)
	if query == "" {
		query = "什么是知识库问答"
	}
	doc1 := strings.TrimSpace(req.Doc1)
	if doc1 == "" {
		doc1 = "知识库问答可以结合企业私有文档，提高回答准确率。"
	}
	doc2 := strings.TrimSpace(req.Doc2)
	if doc2 == "" {
		doc2 = "普通闲聊通常不依赖私有文档检索。"
	}

	data := &types.LLMServiceTestData{
		EmbeddingOk: false,
		RerankOk:    false,
	}

	qc := &utils.QdrantTools{
		HTTPClient: &http.Client{Timeout: 60 * time.Second},
	}
	if embErr := utils.ApplyEmbeddingConfigForClient(&l.svcCtx.Config, qc, clientId); embErr != nil {
		data.EmbeddingError = embErr.Error()
	} else {
		vec, embCallErr := qc.GenerateEmbedding(l.ctx, query)
		if embCallErr != nil {
			data.EmbeddingError = embCallErr.Error()
		} else {
			data.EmbeddingOk = true
			data.EmbeddingDimension = int64(len(vec))
		}
	}

	rerankURL, rerankType, rerankApiKey, rerankModel := utils.ResolveRerankRuntime(&l.svcCtx.Config, clientId)
	if strings.TrimSpace(rerankURL) == "" {
		data.RerankError = "未配置 Rerank URL"
	} else {
		results, rerankErr := utils.RerankByTypeWithAPIKey(l.ctx, rerankURL, rerankApiKey, rerankType, utils.RerankRequest{
			Query:     query,
			Documents: []string{doc1, doc2},
			Model:     rerankModel,
		})
		if rerankErr != nil {
			data.RerankError = rerankErr.Error()
		} else {
			data.RerankOk = true
			scores := make([]float64, 0, len(results))
			top := 0.0
			for _, r := range results {
				scores = append(scores, r.RelevanceScore)
				if r.RelevanceScore > top {
					top = r.RelevanceScore
				}
			}
			data.RerankScores = scores
			data.RerankTopScore = top
		}
	}

	return &types.LLMServiceTestResponse{
		Response: types.Response{
			Code:    response.SuccessCode,
			Message: "success",
		},
		Data: data,
	}, nil
}

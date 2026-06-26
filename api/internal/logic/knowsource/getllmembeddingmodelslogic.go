// 获取 Embedding 模型列表：从 Rag.EmbeddingsUrl 拉取 /v1/models 或 /api/tags，仅返回名称含 embedding 的模型

package knowsource

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"knowsource/api/internal/svc"
	"knowsource/api/internal/types"
	"knowsource/api/internal/utils"
	"knowsource/common/response"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetLLMEmbeddingModelsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetLLMEmbeddingModelsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetLLMEmbeddingModelsLogic {
	return &GetLLMEmbeddingModelsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

type embeddingModelsV1Response struct {
	Data []struct {
		ID string `json:"id"`
	} `json:"data"`
}

type embeddingOllamaTagsResponse struct {
	Models []struct {
		Name string `json:"name"`
	} `json:"models"`
}

func (l *GetLLMEmbeddingModelsLogic) GetLLMEmbeddingModels() (resp *types.LLMChatModelsResponse, err error) {
	clientId, _ := l.ctx.Value("clientId").(string)
	embedBase, _, embeddingsApiKey, _ := utils.ResolveEmbeddingRuntime(&l.svcCtx.Config, clientId)
	if embedBase == "" {
		return &types.LLMChatModelsResponse{
			Response: types.Response{
				Code:    response.SuccessCode,
				Message: "未配置向量化服务（Rag.EmbeddingsUrl）",
			},
			Data: &types.LLMChatModelsData{ModelIds: nil},
		}, nil
	}

	client := &http.Client{Timeout: 8 * time.Second}
	var modelIds []string
	var realurl string

	// GET /v1/models
	u := embedBase + "/v1/models"
	req, errReq := http.NewRequestWithContext(l.ctx, http.MethodGet, u, nil)
	if errReq != nil {
		l.Errorf("创建请求失败: %v", errReq)
		return &types.LLMChatModelsResponse{
			Response: types.Response{Code: response.ServerErrorCode, Message: "创建请求失败", Info: errReq.Error()},
			Data:     &types.LLMChatModelsData{ModelIds: nil},
		}, nil
	}
	if strings.TrimSpace(embeddingsApiKey) != "" {
		req.Header.Set("Authorization", "Bearer "+strings.TrimSpace(embeddingsApiKey))
	}
	res, errDo := client.Do(req)
	if errDo == nil {
		defer res.Body.Close()
		if res.StatusCode == http.StatusOK {
			body, _ := io.ReadAll(res.Body)
			var vllmResp embeddingModelsV1Response
			if json.Unmarshal(body, &vllmResp) == nil {
				for _, m := range vllmResp.Data {
					if m.ID != "" && strings.Contains(strings.ToLower(m.ID), "embedding") {
						modelIds = append(modelIds, m.ID)
					}
				}
				realurl = u
			}
		}
	}

	if len(modelIds) == 0 {
		// 回退到 Ollama：GET /api/tags
		u := embedBase + "/api/tags"
		req, _ := http.NewRequestWithContext(l.ctx, http.MethodGet, u, nil)
		if strings.TrimSpace(embeddingsApiKey) != "" {
			req.Header.Set("Authorization", "Bearer "+strings.TrimSpace(embeddingsApiKey))
		}
		res, errDo := client.Do(req)
		if errDo == nil {
			defer res.Body.Close()
			if res.StatusCode == http.StatusOK {
				body, _ := io.ReadAll(res.Body)
				var ollamaResp embeddingOllamaTagsResponse
				if json.Unmarshal(body, &ollamaResp) == nil {
					for _, m := range ollamaResp.Models {
						if m.Name != "" && strings.Contains(strings.ToLower(m.Name), "embedding") {
							modelIds = append(modelIds, m.Name)
						}
					}
					realurl = u
				}
			}
		}
	}

	return &types.LLMChatModelsResponse{
		Response: types.Response{Code: response.SuccessCode, Message: "success", Info: fmt.Sprintf("embedBase: %s, realurl: %s, modelIds: %v", embedBase, realurl, modelIds)},
		Data:     &types.LLMChatModelsData{ModelIds: modelIds},
	}, nil
}

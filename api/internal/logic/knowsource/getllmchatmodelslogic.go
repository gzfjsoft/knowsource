// 获取对话模型列表：与监控检查中 Chat 使用同一数据源（Llm.CompletionUrl 的 /v1/models 或 Ollama 的 /api/tags）

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

type GetLLMChatModelsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetLLMChatModelsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetLLMChatModelsLogic {
	return &GetLLMChatModelsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// chatModelsV1Response /v1/models 响应结构（OpenAI 兼容）
type chatModelsV1Response struct {
	Data []struct {
		ID string `json:"id"`
	} `json:"data"`
}

// ollamaTagsResponse /api/tags 响应结构
type ollamaTagsResponse struct {
	Models []struct {
		Name string `json:"name"`
	} `json:"models"`
}

func (l *GetLLMChatModelsLogic) GetLLMChatModels() (resp *types.LLMChatModelsResponse, err error) {
	cfg := l.svcCtx.Config
	clientId, _ := l.ctx.Value("clientId").(string)
	chatBase, _, completionApiKey := utils.ResolveCompletionRuntime(&cfg, clientId)

	if chatBase == "" {
		return &types.LLMChatModelsResponse{
			Response: types.Response{
				Code:    response.SuccessCode,
				Message: "未配置对话服务（Llm.CompletionUrl）",
			},
			Data: &types.LLMChatModelsData{ModelIds: nil},
		}, nil
	}

	client := &http.Client{Timeout: 8 * time.Second}
	var modelIds []string
	var realurl string

	if chatBase != "" {
		// GET /v1/models（与监控检查一致）
		u := chatBase + "/v1/models"
		req, errReq := http.NewRequestWithContext(l.ctx, http.MethodGet, u, nil)
		if errReq != nil {
			l.Errorf("创建请求失败: %v", errReq)
			return &types.LLMChatModelsResponse{
				Response: types.Response{Code: response.ServerErrorCode, Message: "创建请求失败", Info: errReq.Error()},
				Data:     &types.LLMChatModelsData{ModelIds: nil},
			}, nil
		}
		if strings.TrimSpace(completionApiKey) != "" {
			req.Header.Set("Authorization", "Bearer "+strings.TrimSpace(completionApiKey))
		}
		res, errDo := client.Do(req)
		if errDo != nil {
			l.Infof("请求 /v1/models 失败，尝试 Ollama: %v", errDo)
		} else {
			defer res.Body.Close()
			if res.StatusCode == http.StatusOK {
				body, _ := io.ReadAll(res.Body)
				var vllmResp chatModelsV1Response
				if json.Unmarshal(body, &vllmResp) == nil {
					for _, m := range vllmResp.Data {
						if m.ID != "" && !strings.Contains(strings.ToLower(m.ID), "embedding") {
							modelIds = append(modelIds, m.ID)
						}
					}
					realurl = u
				}
			}
		}
	}

	if len(modelIds) == 0 && chatBase != "" {
		// 回退到 Ollama：GET /api/tags（同一 base 上）
		u := chatBase + "/api/tags"
		req, errReq := http.NewRequestWithContext(l.ctx, http.MethodGet, u, nil)
		if errReq != nil {
			l.Errorf("创建 Ollama 请求失败: %v", errReq)
			return &types.LLMChatModelsResponse{
				Response: types.Response{Code: response.SuccessCode, Message: "success"},
				Data:     &types.LLMChatModelsData{ModelIds: nil},
			}, nil
		}
		if strings.TrimSpace(completionApiKey) != "" {
			req.Header.Set("Authorization", "Bearer "+strings.TrimSpace(completionApiKey))
		}
		res, errDo := client.Do(req)
		if errDo != nil {
			l.Infof("请求 Ollama /api/tags 失败: %v", errDo)
			return &types.LLMChatModelsResponse{
				Response: types.Response{Code: response.SuccessCode, Message: "success"},
				Data:     &types.LLMChatModelsData{ModelIds: nil},
			}, nil
		}
		defer res.Body.Close()
		if res.StatusCode == http.StatusOK {
			body, _ := io.ReadAll(res.Body)
			var ollamaResp ollamaTagsResponse
			if json.Unmarshal(body, &ollamaResp) == nil {
				for _, m := range ollamaResp.Models {
					if m.Name != "" && !strings.Contains(strings.ToLower(m.Name), "embedding") {
						modelIds = append(modelIds, m.Name)
					}
				}
				realurl = u

			}
		}
	}

	return &types.LLMChatModelsResponse{
		Response: types.Response{Code: response.SuccessCode, Message: "success", Info: fmt.Sprintf("chatBase: %s, realurl: %s, modelIds: %v", chatBase, realurl, modelIds)},
		Data:     &types.LLMChatModelsData{ModelIds: modelIds},
	}, nil
}

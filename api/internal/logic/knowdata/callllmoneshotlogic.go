// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package knowdata

import (
	"context"
	"strings"

	knowsourceLogic "knowsource/api/internal/logic/knowsource"
	"knowsource/api/internal/svc"
	"knowsource/api/internal/types"
	"knowsource/api/internal/utils"
	"knowsource/common/response"

	"github.com/zeromicro/go-zero/core/logx"
)

type CallLLMOneShotLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 调用 LLM 进行单轮对话
func NewCallLLMOneShotLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CallLLMOneShotLogic {
	return &CallLLMOneShotLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CallLLMOneShotLogic) CallLLMOneShot(req *types.CallLLMOneShotRequest) (resp *types.CallLLMOneShotResponse, err error) {
	if req.Prompt == "" {
		return &types.CallLLMOneShotResponse{
			Response: types.Response{
				Code:    response.ParameterErrorCode,
				Message: "提示词不能为空",
			},
		}, nil
	}

	// 从 context 获取 clientId
	clientId, _ := l.ctx.Value("clientId").(string)

	baseURL, completionType, completionApiKey := utils.ResolveCompletionRuntime(&l.svcCtx.Config, clientId)
	ollamaBase := baseURL

	model := ""
	if aiCfg, loadErr := knowsourceLogic.LoadAIConfig(clientId); loadErr == nil && aiCfg != nil && aiCfg.Model != "" {
		model = aiCfg.Model
	}
	if model == "" {
		model = "qwen3:14b"
	}

	var content string
	if completionType == "ollama" && ollamaBase != "" {
		content, err = utils.CallLLMOllamaOneShotWithAPIKey(l.ctx, ollamaBase, completionApiKey, model, req.Prompt, false)
	} else {
		apiUrl := ""
		if strings.TrimSpace(baseURL) != "" {
			apiUrl = strings.TrimSuffix(baseURL, "/") + "/v1/chat/completions"
		}
		if apiUrl == "" {
			return &types.CallLLMOneShotResponse{
				Response: types.Response{
					Code:    response.ServerErrorCode,
					Message: "未配置 LLM 地址（Llm.CompletionUrl）",
				},
			}, nil
		}
		temperature := req.Temperature
		if temperature <= 0 {
			temperature = 0.3
		}
		maxTokens := req.MaxTokens
		content, err = utils.CallLLMOneShotWithAPIKey(l.ctx, apiUrl, completionApiKey, model, req.Prompt, temperature, maxTokens, false)
	}
	if err != nil {
		l.Errorf("调用 LLM 失败: %v", err)
		return &types.CallLLMOneShotResponse{
			Response: types.Response{
				Code:    response.ServerErrorCode,
				Message: "调用 LLM 失败",
				Info:    err.Error(),
			},
		}, nil
	}

	return &types.CallLLMOneShotResponse{
		Response: types.Response{
			Code:    200,
			Message: "success",
		},
		Data: &types.CallLLMOneShotData{
			Content: content,
		},
	}, nil
}

// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package ai

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

type TextConsistencyLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 判断两段文字主要内容是否一致（调用 LLM），供测试脚本使用
func NewTextConsistencyLogic(ctx context.Context, svcCtx *svc.ServiceContext) *TextConsistencyLogic {
	return &TextConsistencyLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *TextConsistencyLogic) TextConsistency(req *types.TextConsistencyRequest) (resp *types.TextConsistencyResponse, err error) {
	// 从 context 获取 clientId，如果没有则使用空字符串
	clientId, _ := l.ctx.Value("clientId").(string)

	text1 := strings.TrimSpace(req.Text1)
	text2 := strings.TrimSpace(req.Text2)
	if text1 == "" || text2 == "" {
		return &types.TextConsistencyResponse{
			Response: types.Response{
				Code:    response.ParameterErrorCode,
				Message: "text1 和 text2 不能为空",
			},
		}, nil
	}
	prompt := req.Prompt + `

文字1：
` + text1 + `

文字2：
` + text2

	baseURL, completionType, completionApiKey := utils.ResolveCompletionRuntime(&l.svcCtx.Config, clientId)
	apiUrl := ""
	if baseURL != "" {
		apiUrl = strings.TrimSuffix(baseURL, "/") + "/v1/chat/completions"
	}
	if apiUrl == "" {
		return &types.TextConsistencyResponse{
			Response: types.Response{
				Code:    response.ServerErrorCode,
				Message: "未配置 LLM 地址（Llm.CompletionUrl）",
			},
		}, nil
	}
	model := ""
	if aiCfg, loadErr := knowsourceLogic.LoadAIConfig(clientId); loadErr == nil && aiCfg != nil && aiCfg.Model != "" {
		model = aiCfg.Model
	}
	model = utils.ResolveChatModel(clientId, model)
	if model == "" {
		return &types.TextConsistencyResponse{
			Response: types.Response{
				Code:    response.ServerErrorCode,
				Message: "未配置 LLM 模型（ai.yaml 中 model 不能为空）",
			},
		}, nil
	}
	var content string
	if completionType == "ollama" {
		content, err = utils.CallLLMOllamaOneShotWithAPIKey(l.ctx, baseURL, completionApiKey, model, prompt, false)
	} else {
		content, err = utils.CallLLMOneShotWithAPIKey(l.ctx, apiUrl, completionApiKey, model, prompt, 0.3, 64, false)
	}
	if err != nil {
		l.Errorf("CallLLMOneShot: %v", err)
		return &types.TextConsistencyResponse{
			Response: types.Response{
				Code:    response.ServerErrorCode,
				Message: "调用 LLM 失败",
				Info:    err.Error(),
			},
		}, nil
	}
	content = strings.ToLower(strings.TrimSpace(content))
	consistent := strings.Contains(content, "yes")
	return &types.TextConsistencyResponse{
		Response: types.Response{
			Code:    200,
			Message: "success",
		},
		Data: types.TextConsistencyData{
			Consistent: consistent,
		},
	}, nil
}

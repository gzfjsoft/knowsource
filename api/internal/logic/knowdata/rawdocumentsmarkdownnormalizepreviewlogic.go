// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package knowdata

import (
	"context"
	"fmt"
	"strings"

	knowsourceLogic "knowsource/api/internal/logic/knowsource"
	"knowsource/api/internal/svc"
	"knowsource/api/internal/types"
	"knowsource/api/internal/utils"
	"knowsource/common/response"
	"knowsource/model"

	"github.com/pkg/errors"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type RawDocumentsMarkdownNormalizePreviewLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// LLM 规范化 Markdown 预览：返回原文与格式化结果，确认后请调用 content/update 保存
func NewRawDocumentsMarkdownNormalizePreviewLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RawDocumentsMarkdownNormalizePreviewLogic {
	return &RawDocumentsMarkdownNormalizePreviewLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

const mdNormalizeSystemHint = `你是 Markdown 语法/排版规范化助手。用户将提供一段知识库文档的 Markdown 原文。

你的任务严格限定为：只修正 Markdown 的语法与排版（标题/空行/列表/引用/代码块围栏/表格/链接/图片等），让它更符合常见 Markdown 规范、渲染更稳定。

强约束（必须遵守）：
1) 正文文字内容必须保持完全不变（逐字不改，不换同义词，不改语序，不改标点），仅允许做 Markdown 语法/排版层面的调整。
2) 仅在“非常明显”的错别字/漏字/多字导致句子明显不通顺且可以确定唯一修正时，才允许改 1-2 个字；除此之外一律不改动文字内容。
3) 不要增删事实信息，不要总结，不要改写，不要补充解释，不要添加原文没有的内容。
4) 必须严格只输出规范化后的完整 Markdown 正文；不要用 markdown 代码块包裹全文；不要添加任何说明、前后缀或对话内容。

--- 原文开始 ---
%s
--- 原文结束 ---`

func stripLLMMarkdownFence(s string) string {
	s = strings.TrimSpace(s)
	if !strings.HasPrefix(s, "```") {
		return s
	}
	rest := s[3:]
	if nl := strings.Index(rest, "\n"); nl >= 0 {
		rest = rest[nl+1:]
	} else {
		return strings.TrimSpace(strings.TrimPrefix(rest, "`"))
	}
	rest = strings.TrimSpace(rest)
	if i := strings.LastIndex(rest, "```"); i >= 0 {
		rest = strings.TrimSpace(rest[:i])
	}
	return rest
}

func (l *RawDocumentsMarkdownNormalizePreviewLogic) RawDocumentsMarkdownNormalizePreview(req *types.RawDocumentsMarkdownNormalizePreviewRequest) (resp *types.RawDocumentsMarkdownNormalizePreviewResponse, err error) {
	clientId, _ := l.ctx.Value("clientId").(string)
	clientId = strings.TrimSpace(clientId)
	if clientId == "" {
		return &types.RawDocumentsMarkdownNormalizePreviewResponse{
			Response: types.Response{
				Code:    response.UnauthorizedCode,
				Message: "clientId不能为空，请重新登录",
			},
		}, nil
	}

	if req.Id <= 0 {
		return &types.RawDocumentsMarkdownNormalizePreviewResponse{
			Response: types.Response{
				Code:    response.ParameterErrorCode,
				Message: "ID 不能为空或无效",
			},
		}, nil
	}

	doc, err := l.svcCtx.RawDocumentsModel.FindOneByClientId(l.ctx, clientId, req.Id)
	if err != nil {
		if err == sqlx.ErrNotFound || errors.Is(err, model.ErrNotFound) {
			return &types.RawDocumentsMarkdownNormalizePreviewResponse{
				Response: types.Response{
					Code:    response.RecordNotExistCode,
					Message: "文档不存在",
				},
			}, nil
		}
		l.Logger.Errorf("查询原始文档失败: %v, ID: %d", err, req.Id)
		return &types.RawDocumentsMarkdownNormalizePreviewResponse{
			Response: types.Response{
				Code:    response.ServerErrorCode,
				Message: "查询失败",
				Info:    err.Error(),
			},
		}, nil
	}

	if doc.IsAudit == 1 {
		return &types.RawDocumentsMarkdownNormalizePreviewResponse{
			Response: types.Response{
				Code:    response.ParameterErrorCode,
				Message: "已审核的文档不能修改内容",
			},
		}, nil
	}

	if doc.IsToMd != 1 {
		return &types.RawDocumentsMarkdownNormalizePreviewResponse{
			Response: types.Response{
				Code:    response.ParameterErrorCode,
				Message: "只有已转换为 Markdown 的文档才能规范化内容",
			},
		}, nil
	}

	source := strings.TrimSpace(req.Content)
	if source == "" {
		source = doc.Content
	}
	originalTrim := strings.TrimSpace(source)
	if originalTrim == "" {
		return &types.RawDocumentsMarkdownNormalizePreviewResponse{
			Response: types.Response{
				Code:    response.ParameterErrorCode,
				Message: "当前文档没有可规范化的 Markdown 内容",
			},
		}, nil
	}

	modelName := ""
	aiCfg, cfgErr := knowsourceLogic.LoadAIConfig(clientId)
	if cfgErr == nil && aiCfg != nil && aiCfg.Model != "" {
		modelName = aiCfg.Model
	}
	modelName = utils.ResolveChatModel(clientId, modelName)
	if modelName == "qwen3" {
		modelName = "Qwen3-0.6B"
	}
	modelName = utils.LLMModelStore.ResolveChatModel(modelName)
	maxTokens := int64(16384)
	temperature := 0.2
	if aiCfg != nil {
		if aiCfg.MaxTokens > 0 {
			maxTokens = aiCfg.MaxTokens
		}
		if aiCfg.Temperature > 0 && aiCfg.Temperature < 1.5 {
			temperature = aiCfg.Temperature
		}
	}
	if maxTokens > 131072 {
		maxTokens = 131072
	}

	prompt := fmt.Sprintf(mdNormalizeSystemHint, source)

	baseURL, completionType, completionApiKey := utils.ResolveCompletionRuntime(&l.svcCtx.Config, clientId)
	ollamaBase := baseURL

	var formatted string
	if completionType == "ollama" && ollamaBase != "" {
		formatted, err = utils.CallLLMOllamaOneShotWithAPIKey(l.ctx, ollamaBase, completionApiKey, modelName, prompt, false)
	} else {
		apiURL := ""
		if baseURL != "" {
			apiURL = strings.TrimSuffix(baseURL, "/") + "/v1/chat/completions"
		}
		if apiURL == "" {
			return &types.RawDocumentsMarkdownNormalizePreviewResponse{
				Response: types.Response{
					Code:    response.ServerErrorCode,
					Message: "未配置 LLM 地址（Llm.CompletionUrl）",
				},
			}, nil
		}
		formatted, err = utils.CallLLMOneShotWithAPIKey(l.ctx, apiURL, completionApiKey, modelName, prompt, temperature, maxTokens, false)
	}
	if err != nil {
		l.Errorf("LLM 规范化失败 id=%d: %v", req.Id, err)
		return &types.RawDocumentsMarkdownNormalizePreviewResponse{
			Response: types.Response{
				Code:    response.ServerErrorCode,
				Message: "LLM 规范化失败",
				Info:    err.Error(),
			},
		}, nil
	}

	formatted = stripLLMMarkdownFence(formatted)
	if formatted == "" {
		return &types.RawDocumentsMarkdownNormalizePreviewResponse{
			Response: types.Response{
				Code:    response.ServerErrorCode,
				Message: "LLM 返回为空，请重试或检查模型配置",
			},
		}, nil
	}

	return &types.RawDocumentsMarkdownNormalizePreviewResponse{
		Response: types.Response{
			Code:    response.SuccessCode,
			Message: "success",
		},
		Data: &types.RawDocumentsMarkdownNormalizePreviewData{
			OriginalContent:  source,
			FormattedContent: formatted,
		},
	}, nil
}

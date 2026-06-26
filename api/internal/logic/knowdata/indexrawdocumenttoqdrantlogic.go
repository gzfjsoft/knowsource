// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package knowdata

import (
	"context"
	"fmt"
	"regexp"
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

type IndexRawDocumentToQdrantLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 将原始文档分块并向量化写入 Qdrant
func NewIndexRawDocumentToQdrantLogic(ctx context.Context, svcCtx *svc.ServiceContext) *IndexRawDocumentToQdrantLogic {
	return &IndexRawDocumentToQdrantLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *IndexRawDocumentToQdrantLogic) IndexRawDocumentToQdrant(req *types.IndexRawDocumentToQdrantRequest) (resp *types.Response, err error) {
	clientId, _ := l.ctx.Value("clientId").(string)
	clientId = strings.TrimSpace(clientId)
	if clientId == "" {
		return &types.Response{
			Code:    response.UnauthorizedCode,
			Message: "clientId不能为空，请重新登录",
		}, nil
	}

	if req.Id <= 0 {
		return &types.Response{
			Code:    response.ServerErrorCode,
			Message: "文档 ID 不能为空或无效",
		}, nil
	}

	doc, err := l.svcCtx.RawDocumentsModel.FindOneByClientId(l.ctx, clientId, req.Id)
	if err != nil {
		if err == sqlx.ErrNotFound || errors.Is(err, model.ErrNotFound) {
			return &types.Response{
				Code:    response.ServerErrorCode,
				Message: "文档不存在",
			}, nil
		}
		l.Errorf("查询原始文档失败: %v", err)
		return &types.Response{
			Code:    response.ServerErrorCode,
			Message: "查询失败",
			Info:    err.Error(),
		}, nil
	}

	if doc.Content == "" {
		return &types.Response{
			Code:    response.ServerErrorCode,
			Message: "文档内容为空，无法索引",
		}, nil
	}

	cfg := l.svcCtx.Config
	if cfg.Qdrant.Host == "" || cfg.Qdrant.Port <= 0 {
		return &types.Response{
			Code:    response.ServerErrorCode,
			Message: "未配置 Qdrant 地址",
		}, nil
	}

	// 1. 生成文档摘要（5000字左右）
	summary := doc.Content
	if len(doc.Content) > 5000 {
		baseURL, completionType, completionApiKey := utils.ResolveCompletionRuntime(&l.svcCtx.Config, clientId)
		apiUrl := ""
		if baseURL != "" {
			apiUrl = strings.TrimSuffix(baseURL, "/") + "/v1/chat/completions"
		}
		model := ""
		if aiCfg, loadErr := knowsourceLogic.LoadAIConfig(clientId); loadErr == nil && aiCfg != nil && aiCfg.Model != "" {
			model = aiCfg.Model
		}
		model = utils.ResolveChatModel(clientId, model)
		if apiUrl != "" && model != "" {
			prompt := "/no_think 请将以下文档内容缩写成约5000字的摘要，保留核心内容和关键信息：\n\n" + doc.Content
			if completionType == "ollama" {
				summarizedContent, err := utils.CallLLMOllamaOneShotWithAPIKey(l.ctx, baseURL, completionApiKey, model, prompt, false)
				if err != nil {
					l.Errorf("调用 LLM 进行文档缩写失败: %v", err)
				} else {
					summary = removeThinkTags(summarizedContent)
					l.Infof("文档缩写成功，原长度: %d, 缩写后长度: %d", len(doc.Content), len(summary))
				}
			} else {
				summarizedContent, err := utils.CallLLMOneShotWithAPIKey(l.ctx, apiUrl, completionApiKey, model, prompt, 0.3, 5000, false)
				if err != nil {
					l.Errorf("调用 LLM 进行文档缩写失败: %v", err)
					// 失败时使用原文
				} else {
					// 移除 <think>...</think> 标签及其内容
					summary = removeThinkTags(summarizedContent)
					l.Infof("文档缩写成功，原长度: %d, 缩写后长度: %d", len(doc.Content), len(summary))
				}
			}
		}
	}

	// 2. 向量化插入摘要到 前缀_clientId_<documentcode>_全文 collection
	prefix := cfg.Qdrant.CollectionPrefix
	summaryCollectionName := utils.FormatCollectionName(prefix, clientId, doc.DocumentCode, true)

	qc, qErr := utils.NewQdrantToolsWithEmbeddingForClient(&cfg, clientId)
	if qErr != nil {
		return &types.Response{
			Code:    response.ServerErrorCode,
			Message: qErr.Error(),
		}, nil
	}

	meta := utils.DocMeta{
		Path:            doc.FilePath,
		FileMD5:         doc.FileMd5,
		FileCreatedTime: float64(doc.CreatedAt.Unix()),
		FileSize:        doc.FileSize,
		Tag:             doc.Tag,
		FileName:        doc.FileName,
		DocType:         doc.DocumentCode,
		Extra: map[string]interface{}{
			"document_code": doc.DocumentCode,
			"rawdoc_id":     doc.Id,
			"is_summary":    true,
		},
	}

	// 向量化插入摘要
	summaryStats, err := utils.IndexDocToQdrantWithStats(l.ctx, summary, meta, summaryCollectionName, qc,
		utils.ChunkModeNone, 0, 0)
	if err != nil {
		l.Errorf("向量化插入摘要失败: %v", err)
		return &types.Response{
			Code:    response.ServerErrorCode,
			Message: "向量化插入摘要失败",
			Info:    err.Error(),
		}, nil
	}

	// 3. 进行原来的拆分写入
	collectionName := req.CollectionName
	if collectionName == "" {
		collectionName = utils.FormatCollectionName(prefix, clientId, doc.DocumentCode, false)
	}

	// 重置 meta 中的 is_summary 标记
	meta.Extra["is_summary"] = false

	mainStats, err := utils.IndexDocToQdrantWithStats(l.ctx, doc.Content, meta, collectionName, qc,
		utils.ChunkModeSuperSmart, l.svcCtx.Config.Document.ChunkSize, l.svcCtx.Config.Document.ChunkOverlap)
	if err != nil {
		l.Errorf("分块并写入 Qdrant 失败: %v", err)
		return &types.Response{
			Code:    response.ServerErrorCode,
			Message: "分块并向量化写入 Qdrant 失败",
			Info:    err.Error(),
		}, nil
	}

	return &types.Response{
		Code:    200,
		Message: "success",
		Info: fmt.Sprintf(
			"qdrant索引完成：摘要%d块(collection=%s,mode=%s)；正文%d块(collection=%s,requested=%s,effective=%s,chunkSize=%d,overlap=%d)",
			summaryStats.ChunkCount, summaryCollectionName, summaryStats.EffectiveMode,
			mainStats.ChunkCount, collectionName, mainStats.RequestedMode, mainStats.EffectiveMode, mainStats.ChunkSize, mainStats.ChunkOverlap,
		),
	}, nil
}

// removeThinkTags 移除字符串中的 <think>...</think> 标签及其内容
func removeThinkTags(content string) string {
	// 正则表达式匹配 <think>...</think> 标签及其内容
	re := regexp.MustCompile(`<think>[\s\S]*?</think>`)
	// 替换为空白字符串
	result := re.ReplaceAllString(content, "")
	// 去除首尾空白
	return strings.TrimSpace(result)
}

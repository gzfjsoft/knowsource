package knowdata

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"time"

	knowsourceLogic "knowsource/api/internal/logic/knowsource"
	"knowsource/api/internal/svc"
	"knowsource/api/internal/types"
	"knowsource/api/internal/utils"
	"knowsource/common/response"
	"knowsource/model"

	"github.com/zeromicro/go-zero/core/logx"
)

type RegenerateAllSummariesLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewRegenerateAllSummariesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RegenerateAllSummariesLogic {
	return &RegenerateAllSummariesLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *RegenerateAllSummariesLogic) RegenerateAllSummaries(req *types.RegenerateAllSummariesRequest) (resp *types.RegenerateAllSummariesResponse, err error) {
	// 设置 60 分钟的超时
	timeoutCtx, cancel := context.WithTimeout(l.ctx, 60*time.Minute)
	defer cancel()

	// 检查 Qdrant 配置
	cfg := l.svcCtx.Config
	if cfg.Qdrant.Host == "" || cfg.Qdrant.Port <= 0 {
		return &types.RegenerateAllSummariesResponse{
			Response: types.Response{Code: response.ServerErrorCode, Message: "未配置 Qdrant 地址"},
		}, nil
	}

	clientId, _ := l.ctx.Value("clientId").(string)
	clientId = strings.TrimSpace(clientId)
	if clientId == "" {
		return &types.RegenerateAllSummariesResponse{
			Response: types.Response{Code: response.UnauthorizedCode, Message: "clientId不能为空，请重新登录"},
		}, nil
	}

	// 获取所有已审核的文档
	documents, err := l.getAuditedDocuments(timeoutCtx, clientId, req.DocumentCode, req.FileIds)
	if err != nil {
		l.Errorf("获取已审核文档失败: %v", err)
		return &types.RegenerateAllSummariesResponse{
			Response: types.Response{Code: response.ServerErrorCode, Message: "获取已审核文档失败", Info: err.Error()},
		}, nil
	}

	total := len(documents)
	processed := 0
	successCount := 0
	failureCount := 0

	// 检查是否是流式请求
	w, ok := l.ctx.Value("http.ResponseWriter").(http.ResponseWriter)
	if !ok {
		w = nil
	}

	// 处理每个文档
	for _, doc := range documents {
		select {
		case <-timeoutCtx.Done():
			return &types.RegenerateAllSummariesResponse{
				Response: types.Response{Code: response.ServerErrorCode, Message: "处理超时"},
				Data: &types.RegenerateAllSummariesData{
					Total:        total,
					Processed:    processed,
					SuccessCount: successCount,
					FailureCount: failureCount,
				},
			}, nil
		default:
			processed++

			// 生成进度更新
			if w != nil {
				l.sendProgress(w, total, processed, successCount, failureCount, fmt.Sprintf("处理文档: %s", doc.FileName))
			}

			// 重新生成概要并更新到 Qdrant
			err := l.regenerateSummary(timeoutCtx, doc, clientId)
			if err != nil {
				l.Errorf("处理文档 %s 失败: %v", doc.FileName, err)
				failureCount++
			} else {
				successCount++
			}
		}
	}

	// 发送最终结果
	if w != nil {
		l.sendProgress(w, total, processed, successCount, failureCount, "处理完成")
	}

	return &types.RegenerateAllSummariesResponse{
		Response: types.Response{Code: response.SuccessCode, Message: "处理完成"},
		Data: &types.RegenerateAllSummariesData{
			Total:        total,
			Processed:    processed,
			SuccessCount: successCount,
			FailureCount: failureCount,
		},
	}, nil
}

// getAuditedDocuments 获取所有已审核的文档
func (l *RegenerateAllSummariesLogic) getAuditedDocuments(ctx context.Context, clientId string, documentCode string, fileIds []int64) ([]*types.RawDocuments, error) {
	var documents []*model.RawDocuments
	var err error

	// 如果指定了文件ID列表，则根据ID获取文档
	if len(fileIds) > 0 {
		documents = make([]*model.RawDocuments, 0, len(fileIds))
		for _, id := range fileIds {
			doc, err := l.svcCtx.RawDocumentsModel.FindOneByClientId(ctx, clientId, id)
			if err != nil {
				l.Errorf("获取文档失败，ID: %d, 错误: %v", id, err)
				continue
			}
			// 只添加已审核的文档
			if doc.IsAudit == 1 {
				documents = append(documents, doc)
			}
		}
	} else {
		// 使用 FindByDocumentCode 方法获取已审核的文档
		documents, err = l.svcCtx.RawDocumentsModel.FindByDocumentCode(ctx, clientId, documentCode, "", "", "1", 0, 10000)
		if err != nil {
			return nil, err
		}
	}

	// 转换为 types.RawDocuments 类型
	var result []*types.RawDocuments
	for _, doc := range documents {
		result = append(result, &types.RawDocuments{
			Id:            doc.Id,
			DocumentCode:  doc.DocumentCode,
			FileMd5:       doc.FileMd5,
			FileName:      doc.FileName,
			FilePath:      doc.FilePath,
			FileSize:      doc.FileSize,
			Content:       doc.Content,
			ContentOrg:    doc.ContentOrg,
			FileList:      doc.FileList,
			Tag:           doc.Tag,
			ZipFileName:   doc.ZipFileName,
			ZipFileSize:   doc.ZipFileSize,
			CreatedAt:     doc.CreatedAt.Unix(),
			UpdatedAt:     doc.UpdatedAt.Unix(),
			IsAudit:       doc.IsAudit,
			IsToMd:        doc.IsToMd,
			IsToAi:        doc.IsToAi,
			UploadUser:    doc.UploadUser,
			UploadEmpcode: doc.UploadEmpcode,
			AuditUser:     doc.AuditUser,
			AuditedAt: func() int64 {
				if doc.AuditedAt.Valid {
					return doc.AuditedAt.Time.Unix()
				}
				return 0
			}(),
			Status: doc.Status,
		})
	}

	return result, nil
}

// regenerateSummary 重新生成文档概要并更新到 Qdrant
func (l *RegenerateAllSummariesLogic) regenerateSummary(ctx context.Context, doc *types.RawDocuments, clientId string) error {
	if doc.Content == "" {
		return fmt.Errorf("文档内容为空: %s", doc.FileName)
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
				summarizedContent, err := utils.CallLLMOllamaOneShotWithAPIKey(ctx, baseURL, completionApiKey, model, prompt, false)
				if err != nil {
					l.Errorf("调用 LLM 进行文档缩写失败: %v", err)
				} else {
					re := regexp.MustCompile(`<think>[\s\S]*?</think>`)
					summary = re.ReplaceAllString(summarizedContent, "")
					summary = strings.TrimSpace(summary)
					l.Infof("文档缩写成功，原长度: %d, 缩写后长度: %d", len(doc.Content), len(summary))
				}
			} else {
				summarizedContent, err := utils.CallLLMOneShotWithAPIKey(ctx, apiUrl, completionApiKey, model, prompt, 0.3, 5000, false)
				if err != nil {
					l.Errorf("调用 LLM 进行文档缩写失败: %v", err)
					// 失败时使用原文
				} else {
					// 移除 <think>...</think> 标签及其内容
					re := regexp.MustCompile(`<think>[\s\S]*?</think>`)
					summary = re.ReplaceAllString(summarizedContent, "")
					summary = strings.TrimSpace(summary)
					l.Infof("文档缩写成功，原长度: %d, 缩写后长度: %d", len(doc.Content), len(summary))
				}
			}
		}
	}

	// 2. 向量化插入摘要到 前缀_clientId_<documentcode>_全文 collection
	prefix := l.svcCtx.Config.Qdrant.CollectionPrefix
	summaryCollectionName := utils.FormatCollectionName(prefix, clientId, doc.DocumentCode, true)

	qc, qErr := utils.NewQdrantToolsWithEmbeddingForClient(&l.svcCtx.Config, clientId)
	if qErr != nil {
		return qErr
	}

	meta := utils.DocMeta{
		Path:            doc.FilePath,
		FileMD5:         doc.FileMd5,
		FileCreatedTime: float64(doc.CreatedAt),
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

	// 先删除旧的概要
	err := qc.DeletePointsByFileName(ctx, summaryCollectionName, doc.FileName)
	if err != nil {
		l.Errorf("删除旧概要失败: %v", err)
		// 继续执行，不阻止新概要的生成
	}

	// 向量化插入新概要
	err = utils.IndexDocToQdrant(ctx, summary, meta, summaryCollectionName, qc,
		utils.ChunkModeNone, 0, 0)
	if err != nil {
		return fmt.Errorf("向量化插入概要失败: %v", err)
	}

	return nil
}

// sendProgress 发送进度更新
func (l *RegenerateAllSummariesLogic) sendProgress(w http.ResponseWriter, total, processed, successCount, failureCount int, message string) {
	progress := map[string]interface{}{
		"total":        total,
		"processed":    processed,
		"successCount": successCount,
		"failureCount": failureCount,
		"message":      message,
		"progress":     float64(processed) / float64(total) * 100,
		"timestamp":    time.Now().Unix(),
	}

	data, err := json.Marshal(progress)
	if err != nil {
		l.Errorf("序列化进度数据失败: %v", err)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	fmt.Fprintf(w, "data: %s\n\n", data)
	if flusher, ok := w.(http.Flusher); ok {
		flusher.Flush()
	}
}

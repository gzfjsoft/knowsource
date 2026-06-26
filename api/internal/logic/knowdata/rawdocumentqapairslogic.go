package knowdata

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"

	"knowsource/api/internal/logic/knowsource"
	"knowsource/api/internal/svc"
	"knowsource/api/internal/utils"
	"knowsource/model"

	"github.com/zeromicro/go-zero/core/logx"
)

type extractedQAPair struct {
	Question string
	Answer   string
}

var (
	qLinePattern = regexp.MustCompile(`(?i)^\s*Q\s*\d*\s*[:：]\s*(.+?)\s*$`)
	aLinePattern = regexp.MustCompile(`(?i)^\s*A\s*\d*\s*[:：]\s*(.+?)\s*$`)
)

const (
	rawDocQaPromptConfigName = "问答提取提示词"
	defaultRawDocQaPromptTpl = "请阅读以下文本片段，站在用户角度提炼该片段能回答的问题及对应答案。请严格按如下格式输出，至少0组，最多5组：\nQ1: xxx\nA1: xxx\nQ2: xxx\nA2: xxx\n不要输出其他解释。\n\n文本片段：\n%s"
)

func qaCollectionName(prefix, clientId, documentCode string) string {
	base := utils.FormatCollectionName(prefix, clientId, documentCode, false)
	return base + "_qa"
}

func buildChunkQAPrompt(promptTpl, chunk string) string {
	tpl := strings.TrimSpace(promptTpl)
	if tpl == "" {
		tpl = defaultRawDocQaPromptTpl
	}
	if strings.Contains(tpl, "%s") {
		return fmt.Sprintf(tpl, chunk)
	}
	return tpl + "\n\n文本片段：\n" + chunk
}

func getOrCreateRawDocQaPromptTemplate(ctx context.Context, svcCtx *svc.ServiceContext, clientId string) (string, error) {
	clientId = strings.TrimSpace(clientId)
	if clientId == "" {
		return defaultRawDocQaPromptTpl, fmt.Errorf("clientId is empty")
	}
	cfg, err := svcCtx.AiConfigModel.FindByNameAndCode(ctx, clientId, rawDocQaPromptConfigName, "")
	if err == nil && cfg != nil && strings.TrimSpace(cfg.Value) != "" {
		return strings.TrimSpace(cfg.Value), nil
	}
	_, insErr := svcCtx.AiConfigModel.Insert(ctx, &model.AiConfig{
		ClientId:     clientId,
		DocumentCode: "",
		Name:         rawDocQaPromptConfigName,
		Value:        defaultRawDocQaPromptTpl,
		CreatedBy:    "system",
	})
	if insErr != nil {
		logx.WithContext(ctx).Errorf("创建默认问答提取提示词失败: %v", insErr)
	}
	return defaultRawDocQaPromptTpl, nil
}

func parseQAPairs(raw string) []extractedQAPair {
	content := strings.TrimSpace(raw)
	if content == "" {
		return nil
	}

	content = strings.ReplaceAll(content, "：", ":")
	content = strings.ReplaceAll(content, "\r\n", "\n")

	pairs := make([]extractedQAPair, 0)

	var currentQ string
	lines := strings.Split(content, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		if qm := qLinePattern.FindStringSubmatch(line); len(qm) > 1 {
			currentQ = strings.TrimSpace(qm[1])
			continue
		}
		if am := aLinePattern.FindStringSubmatch(line); len(am) > 1 {
			ans := strings.TrimSpace(am[1])
			if currentQ != "" && ans != "" {
				pairs = append(pairs, extractedQAPair{Question: currentQ, Answer: ans})
				currentQ = ""
			}
		}
	}
	return pairs
}

// BuildAndStoreRawDocumentQAPairs 审核入库时：按分块抽取问答，写入 MySQL 与 Qdrant。
func BuildAndStoreRawDocumentQAPairs(ctx context.Context, svcCtx *svc.ServiceContext, doc *model.RawDocuments) (int, error) {
	if svcCtx == nil || doc == nil {
		return 0, fmt.Errorf("svcCtx or doc is nil")
	}
	clientId, _ := ctx.Value("clientId").(string)
	clientId = strings.TrimSpace(clientId)
	if clientId == "" {
		return 0, fmt.Errorf("clientId is empty")
	}
	if strings.TrimSpace(doc.Content) == "" {
		return 0, fmt.Errorf("文档内容为空，无法抽取问答")
	}
	logPath := utils.BuildRawDocQaLLMLogPath(clientId, doc.Id, doc.FileName)
	utils.AppendRawDocQaLLMLog(logPath, "start", fmt.Sprintf("rawDocumentId=%d fileName=%s documentCode=%s contentLen=%d", doc.Id, doc.FileName, doc.DocumentCode, len([]rune(doc.Content))))

	baseURL, completionType, completionApiKey := utils.ResolveCompletionRuntime(&svcCtx.Config, clientId)
	apiURL := ""
	if baseURL != "" {
		apiURL = strings.TrimSuffix(baseURL, "/") + "/v1/chat/completions"
	}
	if apiURL == "" {
		utils.AppendRawDocQaLLMLog(logPath, "error", "未配置 LLM 地址")
		return 0, fmt.Errorf("未配置 LLM 地址")
	}
	aiCfg, cfgErr := knowsource.LoadAIConfig(clientId)
	if cfgErr != nil || aiCfg == nil || strings.TrimSpace(aiCfg.Model) == "" {
		utils.AppendRawDocQaLLMLog(logPath, "error", fmt.Sprintf("读取 LLM 配置失败: %v", cfgErr))
		return 0, fmt.Errorf("读取 LLM 配置失败")
	}
	modelName := utils.ResolveChatModel(clientId, strings.TrimSpace(aiCfg.Model))
	promptTpl, promptErr := getOrCreateRawDocQaPromptTemplate(ctx, svcCtx, clientId)
	if promptErr != nil {
		utils.AppendRawDocQaLLMLog(logPath, "error", fmt.Sprintf("读取问答提取提示词失败，使用默认值: %v", promptErr))
	}
	utils.AppendRawDocQaLLMLog(logPath, "config", fmt.Sprintf("apiURL=%s model=%s promptName=%s", apiURL, modelName, rawDocQaPromptConfigName))

	chunks := utils.ChunkDocContent(
		doc.Content,
		utils.ChunkModeSuperSmart,
		svcCtx.Config.Document.ChunkSize,
		svcCtx.Config.Document.ChunkOverlap,
	)
	if len(chunks) == 0 {
		utils.AppendRawDocQaLLMLog(logPath, "error", "文档分块为空，无法抽取问答")
		return 0, fmt.Errorf("文档分块为空，无法抽取问答")
	}
	utils.AppendRawDocQaLLMLog(logPath, "chunk", fmt.Sprintf("totalChunks=%d", len(chunks)))

	qaModel := svcCtx.RawDocumentQaPairsModel
	if qaModel == nil {
		utils.AppendRawDocQaLLMLog(logPath, "error", "RawDocumentQaPairsModel 未初始化")
		return 0, fmt.Errorf("RawDocumentQaPairsModel 未初始化")
	}
	if err := qaModel.DeleteByRawDocumentId(ctx, clientId, doc.Id); err != nil {
		utils.AppendRawDocQaLLMLog(logPath, "error", fmt.Sprintf("清理历史问答失败: %v", err))
		return 0, fmt.Errorf("清理历史问答失败: %w", err)
	}

	qc, qErr := utils.NewQdrantToolsWithEmbeddingForClient(&svcCtx.Config, clientId)
	if qErr != nil {
		utils.AppendRawDocQaLLMLog(logPath, "error", fmt.Sprintf("初始化向量工具失败: %v", qErr))
		return 0, qErr
	}
	collection := qaCollectionName(svcCtx.Config.Qdrant.CollectionPrefix, clientId, doc.DocumentCode)
	if delErr := qc.DeletePointsByFileName(ctx, collection, doc.FileName); delErr != nil {
		logx.WithContext(ctx).Infof("qa collection cleanup skipped, collection=%s err=%v", collection, delErr)
	}

	totalInserted := 0
	collectionReady := false

	for i, chunk := range chunks {
		prompt := buildChunkQAPrompt(promptTpl, chunk)
		logx.WithContext(ctx).Infof("qa-extract call llm chunk=%d/%d rawDocId=%d model=%s", i+1, len(chunks), doc.Id, modelName)
		utils.AppendRawDocQaLLMLog(logPath, fmt.Sprintf("chunk-%d-request", i+1), fmt.Sprintf("chunkLen=%d\nprompt=\n%s", len([]rune(chunk)), prompt))
		callStart := time.Now()
		var resp string
		var err error
		if completionType == "ollama" {
			resp, err = utils.CallLLMOllamaOneShotWithAPIKey(ctx, baseURL, completionApiKey, modelName, prompt, false)
		} else {
			resp, err = utils.CallLLMOneShotWithAPIKey(ctx, apiURL, completionApiKey, modelName, prompt, 0.2, 1200, false)
		}
		if err != nil {
			logx.WithContext(ctx).Errorf("chunk %d qa extract failed: %v", i+1, err)
			utils.AppendRawDocQaLLMLog(logPath, fmt.Sprintf("chunk-%d-error", i+1), fmt.Sprintf("costMs=%d err=%v", time.Since(callStart).Milliseconds(), err))
			continue
		}
		utils.AppendRawDocQaLLMLog(logPath, fmt.Sprintf("chunk-%d-response", i+1), fmt.Sprintf("costMs=%d\nresponse=\n%s", time.Since(callStart).Milliseconds(), resp))

		qaPairs := parseQAPairs(resp)
		utils.AppendRawDocQaLLMLog(logPath, fmt.Sprintf("chunk-%d-parse", i+1), fmt.Sprintf("parsedPairs=%d", len(qaPairs)))
		if len(qaPairs) == 0 {
			continue
		}

		for _, qa := range qaPairs {
			question := strings.TrimSpace(qa.Question)
			answer := strings.TrimSpace(qa.Answer)
			if question == "" || answer == "" {
				continue
			}

			vec, embErr := qc.GenerateEmbedding(ctx, question)
			if embErr != nil {
				logx.WithContext(ctx).Errorf("qa embedding failed: %v", embErr)
				utils.AppendRawDocQaLLMLog(logPath, fmt.Sprintf("chunk-%d-embedding-error", i+1), fmt.Sprintf("question=%s err=%v", question, embErr))
				continue
			}

			if !collectionReady {
				exists, exErr := qc.CollectionExists(ctx, collection)
				if exErr != nil {
					utils.AppendRawDocQaLLMLog(logPath, fmt.Sprintf("chunk-%d-collection-error", i+1), fmt.Sprintf("检查集合失败: %v", exErr))
					return totalInserted, fmt.Errorf("检查问答集合失败: %w", exErr)
				}
				if !exists {
					if cErr := qc.CreateCollection(ctx, collection, len(vec), "Dot"); cErr != nil {
						utils.AppendRawDocQaLLMLog(logPath, fmt.Sprintf("chunk-%d-collection-error", i+1), fmt.Sprintf("创建集合失败: %v", cErr))
						return totalInserted, fmt.Errorf("创建问答集合失败: %w", cErr)
					}
					_ = qc.CreatePayloadIndex(ctx, collection, "metadata.file_name", "keyword")
				}
				collectionReady = true
			}

			pointID := utils.GenerateUUID()
			payload := map[string]interface{}{
				"metadata": map[string]interface{}{
					"client_id":     clientId,
					"rawdoc_id":     doc.Id,
					"document_code": doc.DocumentCode,
					"file_name":     doc.FileName,
					"chunk_index":   i + 1,
					"is_qa":         true,
					"tag":           doc.Tag,
				},
				"page_content": question,
				"answer":       answer,
			}
			point := utils.Point{
				ID:      pointID,
				Vector:  vec,
				Payload: payload,
			}
			if upErr := qc.UpsertPoints(ctx, collection, []utils.Point{point}); upErr != nil {
				logx.WithContext(ctx).Errorf("upsert qa point failed: %v", upErr)
				utils.AppendRawDocQaLLMLog(logPath, fmt.Sprintf("chunk-%d-upsert-error", i+1), fmt.Sprintf("question=%s err=%v", question, upErr))
				continue
			}

			if _, insErr := qaModel.Insert(ctx, &model.RawDocumentQaPairs{
				ClientId:      clientId,
				RawDocumentId: doc.Id,
				DocumentCode:  doc.DocumentCode,
				FileName:      doc.FileName,
				ChunkIndex:    int64(i + 1),
				Question:      question,
				Answer:        answer,
				QdrantPointId: pointID,
			}); insErr != nil {
				logx.WithContext(ctx).Errorf("insert raw_document_qa_pairs failed: %v", insErr)
				utils.AppendRawDocQaLLMLog(logPath, fmt.Sprintf("chunk-%d-insert-error", i+1), fmt.Sprintf("question=%s err=%v", question, insErr))
				continue
			}
			utils.AppendRawDocQaLLMLog(logPath, fmt.Sprintf("chunk-%d-inserted", i+1), fmt.Sprintf("question=%s answer=%s", question, answer))
			totalInserted++
		}
	}

	if totalInserted == 0 {
		utils.AppendRawDocQaLLMLog(logPath, "finish", "未抽取到有效问答")
		return 0, fmt.Errorf("未抽取到有效问答")
	}
	utils.AppendRawDocQaLLMLog(logPath, "finish", fmt.Sprintf("抽取完成 totalInserted=%d", totalInserted))
	return totalInserted, nil
}

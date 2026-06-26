package ai

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	aiLogic "knowsource/api/internal/logic/ai"
	knowsourceLogic "knowsource/api/internal/logic/knowsource"
	"knowsource/api/internal/svc"
	"knowsource/api/internal/types"
	"knowsource/api/internal/utils"
	"knowsource/model"

	"github.com/google/uuid"
	"github.com/zeromicro/go-zero/core/logx"
)

// LLM 上游 HTTP 超时：非流式整段请求/响应；流式为从开始读到流结束的上限（与 context 一并约束）。
const (
	llmHTTPNonStreamTimeout = 10 * time.Minute
	llmHTTPStreamTimeout    = 15 * time.Minute
)

func newLLMHTTPClient(nonStream bool) *http.Client {
	if nonStream {
		return &http.Client{Timeout: llmHTTPNonStreamTimeout}
	}
	return &http.Client{Timeout: llmHTTPStreamTimeout}
}

// AIOptions holds tunable parameters for AI calls
type AIOptions struct {
	Temperature float64 `json:"temperature,omitempty"`
	TopK        int64   `json:"top_k,omitempty"`
	TopP        float64 `json:"top_p,omitempty"`
	MaxTokens   int64   `json:"max_tokens,omitempty"`
}

// AIRequest represents the request structure for the AI API
type AIRequest struct {
	Model       string    `json:"model"`
	Messages    []Message `json:"messages"`
	Stream      bool      `json:"stream"`
	Think       bool      `json:"think"`
	Temperature float64   `json:"temperature"`
	TopK        int64     `json:"top_k"`
	TopP        float64   `json:"top_p"`
	MaxTokens   int64     `json:"max_tokens"`
	AIOptions   AIOptions `json:"ai_options,omitempty"`
	// RepeatPenalty float64   `json:"repeatPenalty"`
}

// OpenAIRequest represents the OpenAI-compatible request structure for vLLM / OpenAI-style chat APIs
type OpenAIRequest struct {
	Model       string                 `json:"model,omitempty"`
	Messages    []Message              `json:"messages"`
	Stream      bool                   `json:"stream"`
	Temperature float64                `json:"temperature,omitempty"`
	TopP        float64                `json:"top_p,omitempty"`
	MaxTokens   int64                  `json:"max_tokens,omitempty"`
	ExtraBody   map[string]interface{} `json:"extra_body,omitempty"`
}

// Message represents a message in the conversation
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// AIResponse represents the response from the AI API
type AIResponse struct {
	Model     string `json:"model"`
	CreatedAt string `json:"created_at"`
	Message   struct {
		Role    string `json:"role"`
		Content string `json:"content"`
	} `json:"message"`
	DoneReason string `json:"done_reason"`
	Done       bool   `json:"done"`
}

// ChatRequest represents the request from the client
type ChatRequest struct {
	Message string `json:"message"`
	Session string `json:"session,omitempty"`
	Think   bool   `json:"think,omitempty"`
}

// ChatResponse represents the response to the client
type ChatResponse struct {
	Response string `json:"response"`
	Session  string `json:"session,omitempty"`
	Error    string `json:"error,omitempty"`
}

// rawDocMeta 流式响应首行 meta 中的参考资料 id/文件名，供前端解析
type rawDocMeta struct {
	Id       int64  `json:"id"`
	FileName string `json:"fileName"`
}

// timingReader 包装下游流，在 Read 时统计首 token 与整体流式耗时（基于 callStart）。
type timingReader struct {
	r             io.Reader
	callStart     time.Time
	firstTokenMs  *int64
	totalStreamMs *int64
	firstSet      bool
}

func (tr *timingReader) Read(p []byte) (int, error) {
	n, err := tr.r.Read(p)
	if n > 0 && tr.firstTokenMs != nil && !tr.callStart.IsZero() && !tr.firstSet {
		*tr.firstTokenMs = time.Since(tr.callStart).Milliseconds()
		tr.firstSet = true
	}
	if err == io.EOF && tr.totalStreamMs != nil && !tr.callStart.IsZero() {
		*tr.totalStreamMs = time.Since(tr.callStart).Milliseconds()
	}
	return n, err
}

// writeStreamMetaAndCopy 先透传 LLM 流，并捕获完整流内容（不追加任何额外 SSE 事件）。
// 返回 injectAndCopy 的捕获内容。
// stats 为 map[string]interface{}，其中可包含 fullTextMs/mainSearchMs/subSearchMs/firstTokenMs/totalStreamMs/modelName 等。
func writeStreamMetaAndCopy(
	w http.ResponseWriter,
	flusher http.Flusher,
	fileinfos []string,
	rawDocs []rawDocMeta,
	body io.Reader,
	stats map[string]interface{},
	callStart time.Time,
	firstTokenMs *int64,
	totalStreamMs *int64,
) (infoString string, captured string) {
	// 先流式透传 LLM 返回，并在 timingReader 内统计首字/总耗时
	tr := &timingReader{
		r:             body,
		callStart:     callStart,
		firstTokenMs:  firstTokenMs,
		totalStreamMs: totalStreamMs,
	}
	// 不在 injectAndCopy 里追加任何额外行
	captured = injectAndCopy(w, tr, nil)

	// 补全 stats 中的首字/总耗时
	if stats != nil {
		if firstTokenMs != nil {
			stats["firstTokenMs"] = *firstTokenMs
		}
		if totalStreamMs != nil {
			stats["totalStreamMs"] = *totalStreamMs
		}
	}

	return infoString, captured
}

// writeStreamMeta 在流式响应结束后，追加一条 meta SSE 事件给前端。
// 注意：这里的 meta 不是上游 LLM 的 chunk，而是服务端自定义的“检索/归因信息”。
func writeStreamMeta(
	w http.ResponseWriter,
	flusher http.Flusher,
	fileinfos []string,
	rawDocs []rawDocMeta,
	stats map[string]interface{},
	attributedFiles []string,
) (metaLine string) {
	fileinfos = dedupeStrings(fileinfos)
	attributedFiles = dedupeStrings(attributedFiles)
	type metaStruct struct {
		FilesInfos       []string     `json:"filesinfos"`
		RawDocs          []rawDocMeta `json:"rawdocs,omitempty"`
		Stats            interface{}  `json:"stats,omitempty"`
		AttributedFiles  []string     `json:"attributedFiles,omitempty"`
		AttributionNotes string       `json:"attributionNotes,omitempty"`
	}
	meta := metaStruct{
		FilesInfos:      fileinfos,
		RawDocs:         rawDocs,
		Stats:           stats,
		AttributedFiles: attributedFiles,
	}
	if len(attributedFiles) > 0 {
		meta.AttributionNotes = "attributedFiles 来自答案 vs 文件名 rerank (阈值0.8)"
	}
	if metaBytes, err := json.Marshal(meta); err == nil {
		metaLine = fmt.Sprintf("data: %s\n", string(metaBytes))
		_, _ = w.Write([]byte(metaLine))
		if flusher != nil {
			flusher.Flush()
		}
	}
	return metaLine
}

// AskAIOllama 调用 Llm.CompletionUrl（OpenAI 兼容 /v1/chat/completions），进行一次性非流式对话
func AskAIOllama(ctx context.Context, svcCtx *svc.ServiceContext, session *model.AiSessions, question string, model string, aiOptions AIOptions, think bool) (string, error) {
	return AskAIVllm(ctx, svcCtx, session, question, model, aiOptions, think)
}

// AskAIVllm 调用 vLLM（OpenAI Chat Completions 风格）或回退到 Ollama，进行一次性非流式对话。
// 与 AskAIOllama 类似，但优先走 vLLM 的 OpenAI 兼容接口。
func AskAIVllm(ctx context.Context, svcCtx *svc.ServiceContext, session *model.AiSessions, question string, model string, aiOptions AIOptions, think bool) (string, error) {
	messages := []Message{
		{
			Role:    "user",
			Content: strings.TrimSpace(question),
		},
	}

	openAIReq := OpenAIRequest{
		Model:       model,
		Messages:    messages,
		Stream:      false,
		Temperature: aiOptions.Temperature,
		TopP:        aiOptions.TopP,
		MaxTokens:   aiOptions.MaxTokens,
	}

	// For Qwen3.5 models in no-think mode, explicitly disable thinking via extra_body
	if !think {
		lowerModel := strings.ToLower(model)
		if strings.Contains(lowerModel, "qwen3.5") {
			openAIReq.ExtraBody = map[string]interface{}{
				"enable_thinking": false,
			}
		}
	}

	clientId, _ := ctx.Value("clientId").(string)
	baseURL, _, completionApiKey := utils.ResolveCompletionRuntime(&svcCtx.Config, clientId)
	apiUrl := ""
	if baseURL != "" {
		apiUrl = strings.TrimSuffix(baseURL, "/") + "/v1/chat/completions"
	}
	if apiUrl == "" {
		return "", fmt.Errorf("未配置 LLM 地址（Llm.CompletionUrl）")
	}

	jsonData, err := json.Marshal(openAIReq)
	if err != nil {
		return "", fmt.Errorf("failed to marshal vLLM request: %v", err)
	}

	logx.Debugf("call vLLM/Ollama API url=%s bodyLen=%d", apiUrl, len(jsonData))

	callCtx, cancel := context.WithTimeout(ctx, llmHTTPNonStreamTimeout)
	defer cancel()
	req, err := http.NewRequestWithContext(callCtx, "POST", apiUrl, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create vLLM request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	if strings.TrimSpace(completionApiKey) != "" {
		req.Header.Set("Authorization", "Bearer "+strings.TrimSpace(completionApiKey))
	}

	resp, err := newLLMHTTPClient(true).Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to contact vLLM/Ollama API: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read vLLM/Ollama response body: %v", err)
	}

	return string(body), nil
}

// runRAGScript loads the RAG script, runs it with the given request params and fullTextMdStrings,
// and returns fileinfos, the possibly updated message (with RAG context), or an error.
// If the RAG script is missing, err is non-nil and the caller should respond with HTTP 500.
//  fileinfos, chatReq.Message, err = runRAGScript(r.Context(), svcCtx, chatReq.Message, chatReq.Keys, chatReq.DocumentCode, chatReq.Tags, FullTextMdStrings, AI_FORMAT_STRING)

func runRAGScript(ctx context.Context, svcCtx *svc.ServiceContext, message, keys, documentCode string, tags []string, fullTextMdStrings []string, aiFormatString []byte) (fileinfos []string, newMessage string, err error) {
	var ragScript []byte
	clientId, _ := ctx.Value("clientId").(string)
	ragJsConfig, configErr := svcCtx.AiConfigModel.FindByNameAndCode(ctx, clientId, "rag.js", "")
	if configErr == nil && ragJsConfig != nil {
		ragScript = []byte(ragJsConfig.Value)
	} else {
		logx.Debugf("从数据库中获取 rag.js 失败，回退本地文件: %v", configErr)
		ragScript, err = os.ReadFile("./rag.js")
		if err != nil {
			return nil, "", err
		}
	}
	if len(ragScript) == 0 || strings.TrimSpace(string(ragScript)) == "" {
		return nil, "", errors.New("找不到 RAG 脚本")
	}

	toJsString := func(s string) string {
		b, _ := json.Marshal(s)
		return string(b)
	}
	toJsArray := func(s []string) string {
		b, _ := json.Marshal(s)
		return string(b)
	}

	stringContent := strings.ReplaceAll(string(ragScript), "{{query}}", toJsString(message))
	stringContent = strings.ReplaceAll(stringContent, "{{keys}}", toJsString(keys))
	stringContent = strings.ReplaceAll(stringContent, "{{documentCode}}", toJsString(documentCode))
	stringContent = strings.ReplaceAll(stringContent, "{{tags}}", toJsArray(tags))
	stringContent = strings.ReplaceAll(stringContent, "{{collectionPrefix}}", toJsString(svcCtx.Config.Qdrant.CollectionPrefix))
	stringContent = strings.ReplaceAll(stringContent, "{{RAGURL}}", toJsString(svcCtx.Config.RAGURL))

	result := utils.GojsCall(stringContent)

	if result["success"] != true {
		logx.Debugf("RAG 检索失败: %v", result["error"])
		return nil, message, nil
	}

	if result["jsonfileinfo"] != nil {
		if arr, ok := result["jsonfileinfo"].([]interface{}); ok {
			for _, v := range arr {
				if s, ok := v.(string); ok {
					fileinfos = append(fileinfos, s)
				}
			}
		}
	}

	jsResponse, _ := result["response"].(string)
	for _, v := range fullTextMdStrings {
		jsResponse += "\n\n" + v
	}

	if jsResponse != "" {
		logx.Debugf("runRAGScript jsResponse len=%d", len(jsResponse))
		newMessage = fmt.Sprintf(string(aiFormatString), jsResponse, message)
		return fileinfos, newMessage, nil
	}

	return fileinfos, message, nil
}

// streamParseResult 流式解析结果：合并后的回复内容 + 从 stream 中解析出的 model（若有）
type streamParseResult struct {
	Content string
	Reason  string
	Model   string
}

func appendIfNonEmpty(sb *strings.Builder, s string) {
	if strings.TrimSpace(s) != "" {
		sb.WriteString(s)
	}
}

func readChunkString(v interface{}) string {
	switch vv := v.(type) {
	case string:
		return vv
	case []interface{}:
		var sb strings.Builder
		for _, item := range vv {
			switch iv := item.(type) {
			case string:
				sb.WriteString(iv)
			case map[string]interface{}:
				if t, ok := iv["text"].(string); ok {
					sb.WriteString(t)
				} else if t, ok := iv["content"].(string); ok {
					sb.WriteString(t)
				}
			}
		}
		return sb.String()
	default:
		return ""
	}
}

func appendReasoningFromMap(sb *strings.Builder, m map[string]interface{}) {
	if m == nil {
		return
	}
	appendIfNonEmpty(sb, readChunkString(m["reasoning"]))
	appendIfNonEmpty(sb, readChunkString(m["thinking"]))
	appendIfNonEmpty(sb, readChunkString(m["reasoning_content"]))
	appendIfNonEmpty(sb, readChunkString(m["reasoningContent"]))
	appendIfNonEmpty(sb, readChunkString(m["thinking_content"]))
	appendIfNonEmpty(sb, readChunkString(m["thinkingContent"]))
}

// parseStreamResponse parses multiple JSON objects from stream response and merges their content;
// also returns model from the first chunk that has it (e.g. OpenAI "model":"Qwen3-0.6B").
// Supports both Ollama and vLLM (OpenAI-compatible) formats.
func parseStreamResponse(response string) streamParseResult {
	var mergedContent strings.Builder
	var mergedReason strings.Builder
	var streamModel string

	lines := strings.Split(response, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		isDataPrefix := false
		if strings.HasPrefix(line, "data: ") {
			line = strings.TrimPrefix(line, "data: ")
			isDataPrefix = true
			if line == "[DONE]" {
				continue
			}
		}

		if !strings.HasPrefix(line, "{") {
			continue
		}

		if isDataPrefix {
			var openAIChunk map[string]interface{}
			if err := json.Unmarshal([]byte(line), &openAIChunk); err == nil {
				if streamModel == "" {
					if m, ok := openAIChunk["model"].(string); ok && m != "" {
						streamModel = m
					}
				}
				if choices, ok := openAIChunk["choices"].([]interface{}); ok {
					for _, choice := range choices {
						if choiceMap, ok := choice.(map[string]interface{}); ok {
							if delta, ok := choiceMap["delta"].(map[string]interface{}); ok {
								if content, ok := delta["content"].(string); ok && content != "" {
									mergedContent.WriteString(content)
								}
								appendReasoningFromMap(&mergedReason, delta)
							} else if message, ok := choiceMap["message"].(map[string]interface{}); ok {
								if content, ok := message["content"].(string); ok && content != "" {
									mergedContent.WriteString(content)
								}
								appendReasoningFromMap(&mergedReason, message)
							}
						}
					}
					continue
				}
			}
		}

		var ollamaResp types.OllamaResponse
		if err := json.Unmarshal([]byte(line), &ollamaResp); err == nil {
			if ollamaResp.Message.Content != "" {
				mergedContent.WriteString(ollamaResp.Message.Content)
			}
			if ollamaResp.Thinking != "" {
				mergedReason.WriteString(ollamaResp.Thinking)
			}
			var generic map[string]interface{}
			if gErr := json.Unmarshal([]byte(line), &generic); gErr == nil {
				if msgMap, ok := generic["message"].(map[string]interface{}); ok {
					appendReasoningFromMap(&mergedReason, msgMap)
				}
				appendReasoningFromMap(&mergedReason, generic)
			}
			continue
		}

		var vllmResp types.VLLMResponse
		if err := json.Unmarshal([]byte(line), &vllmResp); err == nil {
			for _, choice := range vllmResp.Choices {
				if choice.Message.Content != "" {
					mergedContent.WriteString(choice.Message.Content)
				}
			}
			continue
		}

		var openAIChunk map[string]interface{}
		if err := json.Unmarshal([]byte(line), &openAIChunk); err == nil {
			if streamModel == "" {
				if m, ok := openAIChunk["model"].(string); ok && m != "" {
					streamModel = m
				}
			}
			if choices, ok := openAIChunk["choices"].([]interface{}); ok {
				for _, choice := range choices {
					if choiceMap, ok := choice.(map[string]interface{}); ok {
						if delta, ok := choiceMap["delta"].(map[string]interface{}); ok {
							if content, ok := delta["content"].(string); ok && content != "" {
								mergedContent.WriteString(content)
							}
							appendReasoningFromMap(&mergedReason, delta)
						} else if message, ok := choiceMap["message"].(map[string]interface{}); ok {
							if content, ok := message["content"].(string); ok && content != "" {
								mergedContent.WriteString(content)
							}
							appendReasoningFromMap(&mergedReason, message)
						}
					}
				}
				continue
			}
		}

		logx.Debugf("parseStreamResponse skip line len=%d", len(line))
	}

	result := mergedContent.String()
	if result == "" {
		logx.Debugf("parseStreamResponse: no content extracted from stream response")
	}

	reason := mergedReason.String()
	if strings.TrimSpace(reason) == "" {
		if thinkMatch := thinkBlockRE.FindStringSubmatch(result); len(thinkMatch) > 1 {
			reason = strings.TrimSpace(thinkMatch[1])
		}
	}

	return streamParseResult{Content: result, Reason: reason, Model: streamModel}
}

// getOrCreateSessionFromDB creates or retrieves a session from the database
func getOrCreateSessionFromDB(ctx context.Context, svcCtx *svc.ServiceContext, sessionUUID string, userID int64, modelname string, categoryId int64, keys string, empCode string, documentTypeCode string, tags string) (*model.AiSessions, error) {
	clientId, _ := ctx.Value("clientId").(string)
	if strings.TrimSpace(clientId) == "" {
		return nil, errors.New("clientId不能为空")
	}

	if sessionUUID == "" {
		// Create new session
		sessionUUID = uuid.New().String()
		session := &model.AiSessions{
			ClientId:         clientId,
			SessionUuid:      sessionUUID,
			UserId:           userID,
			CategoryId:       categoryId,
			EmpCode:          empCode,
			Keys:             keys,
			Tags:             tags,
			DocumentTypeCode: documentTypeCode,
			SessionTitle:     "新对话",
			SessionStatus:    "active",
			Model:            modelname,
			LastMessageTime:  sql.NullTime{Time: time.Now(), Valid: true},
			MessageCount:     0,
		}

		_, err := svcCtx.AiSessionsModel.Insert(ctx, session)
		if err != nil {
			return nil, err
		}

		// Get the created session with ID
		session, err = svcCtx.AiSessionsModel.FindOneByClientIdSessionUuid(ctx, clientId, sessionUUID)
		if err != nil {
			return nil, err
		}

		return session, nil
	} else {
		// Find existing session
		session, err := svcCtx.AiSessionsModel.FindOneByClientIdSessionUuid(ctx, clientId, sessionUUID)
		if err != nil {
			// If session not found, create new one
			sessionUUID = uuid.New().String()
			session := &model.AiSessions{
				ClientId:         clientId,
				SessionUuid:      sessionUUID,
				UserId:           userID,
				CategoryId:       categoryId,
				Keys:             keys,
				DocumentTypeCode: documentTypeCode,
				EmpCode:          empCode,
				SessionTitle:     "新对话",
				SessionStatus:    "active",
				Model:            modelname,
				Tags:             tags,
				LastMessageTime:  sql.NullTime{Time: time.Now(), Valid: true},
				MessageCount:     0,
			}

			_, err := svcCtx.AiSessionsModel.Insert(ctx, session)
			if err != nil {
				return nil, err
			}

			// Get the created session with ID
			session, err = svcCtx.AiSessionsModel.FindOneByClientIdSessionUuid(ctx, clientId, sessionUUID)
			if err != nil {
				return nil, err
			}

			return session, nil
		}

		// Update session last message time
		session.LastMessageTime = sql.NullTime{Time: time.Now(), Valid: true}
		session.MessageCount++
		err = svcCtx.AiSessionsModel.Update(ctx, session)
		if err != nil {
			logx.Errorf("更新会话失败: %v", err)
		}

		return session, nil
	}
}

// dedupeStrings returns a new slice with duplicate strings removed, preserving first occurrence order.
func dedupeStrings(ss []string) []string {
	if len(ss) == 0 {
		return ss
	}
	seen := make(map[string]struct{})
	out := make([]string, 0, len(ss))
	for _, s := range ss {
		s = strings.TrimSpace(s)
		if s == "" {
			continue
		}
		if _, ok := seen[s]; ok {
			continue
		}
		seen[s] = struct{}{}
		out = append(out, s)
	}
	return out
}

// saves a message to the database
func saveMessageToDB(ctx context.Context, svcCtx *svc.ServiceContext, sessionID uint64, role, content, modelname, context, thinking string, fileinfos []string) error {
	clientId, _ := ctx.Value("clientId").(string)
	if strings.TrimSpace(clientId) == "" {
		return errors.New("clientId不能为空")
	}

	filesJSON := ""
	if len(fileinfos) > 0 {
		// 去重，避免同一文件在数据库中重复存储多次
		fileinfos = dedupeStrings(fileinfos)
		if b, err := json.Marshal(fileinfos); err == nil {
			filesJSON = string(b)
		}
	}
	message := &model.AiMessages{
		ClientId:      clientId,
		SessionId:     sessionID,
		MessageUuid:   uuid.New().String(),
		Role:          role,
		Content:       strings.TrimSpace(content),
		Context:       strings.TrimSpace(context),
		Model:         modelname,
		Files:         filesJSON,
		CreatedAtUnix: time.Now().Unix(),
	}
	if strings.TrimSpace(thinking) != "" {
		message.Thinking = sql.NullString{String: strings.TrimSpace(thinking), Valid: true}
	}

	_, err := svcCtx.AiMessagesModel.Insert(ctx, message)
	if err != nil {
		logx.Errorf("保存消息失败: %v", err)
		return err
	}

	return nil
}

// sessionLogDir 存放同一 session 对话日志的目录
const sessionLogDir = "ai_session_logs"

// sanitizeForFilename 将字符串中的非法文件名字符替换为下划线
func sanitizeForFilename(s string) string {
	if s == "" {
		return "unknown"
	}
	re := regexp.MustCompile(`[<>:"/\\|?*\s]+`)
	return strings.TrimRight(re.ReplaceAllString(s, "_"), "_.")
}

// sessionLogPath 返回按 clientid、session 与用户命名的日志文件路径
func sessionLogPath(clientId string, sessionID uint64, userName string) string {
	// 创建 clientid 目录
	clientLogDir := filepath.Join(sessionLogDir, clientId)
	_ = os.MkdirAll(clientLogDir, 0755)
	name := "session_" + strconv.FormatUint(sessionID, 10) + "_" + sanitizeForFilename(userName) + ".log.txt"
	return filepath.Join(clientLogDir, name)
}

// appendSessionLog 向同一 session 的日志文件追加一条记录（同一对话 session 同一文件，后续问答追加）
func appendSessionLog(filePath, role, content string) {
	f, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		logx.Errorf("打开会话日志文件 %s 失败: %v", filePath, err)
		return
	}
	defer f.Close()
	ts := time.Now().Format("2006-01-02 15:04:05")
	block := fmt.Sprintf("\n[%s] [%s]\n%s\n", ts, role, strings.TrimSpace(content))
	if _, err := f.WriteString(block); err != nil {
		logx.Errorf("追加会话日志失败: %v", err)
	}
}
func returnErrorJson(w http.ResponseWriter, status int, error string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(ChatResponse{
		Error: error,
	})
}

func AISessionChatHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Enable CORS
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		// Parse request
		var chatReq types.AISessionChatRequest
		if err := json.NewDecoder(r.Body).Decode(&chatReq); err != nil {
			returnErrorJson(w, http.StatusBadRequest, "Invalid request format")
			return
		}

		if chatReq.Message == "" {
			returnErrorJson(w, http.StatusBadRequest, "Message is required")
			return
		}

		chatReq.Message = strings.TrimSpace(chatReq.Message)

		originalUserMessage := chatReq.Message

		// Get user ID and user name from context (set by auth middleware)
		empCode, ok := r.Context().Value("empCode").(string)
		if !ok {
			returnErrorJson(w, http.StatusUnauthorized, "User not authenticated")
			return
		}
		userNameCn, ok := r.Context().Value("userName").(string) // 用户中文名
		if !ok {
			userNameCn = empCode
		}

		// 读取临时对话文档缓存（上传的 txt/docx/pdf 内容），合并进本条消息后清除
		hasUploadedDocs := false
		if svcCtx.RedisClient != nil && empCode != "" {
			cachedList, getErr := utils.AIChatDocCacheGet(svcCtx.RedisClient, empCode)
			if getErr == nil && len(cachedList) > 0 {
				hasUploadedDocs = true
				prefix := utils.BuildContextFromCachedDocs(cachedList)
				chatReq.Message = prefix + chatReq.Message
				_ = utils.AIChatDocCacheClear(svcCtx.RedisClient, empCode)
			}
		}

		userIDValue := r.Context().Value("userId")

		if userIDValue == nil {
			returnErrorJson(w, http.StatusUnauthorized, "User not authenticated")
			return
		}

		userID, ok := userIDValue.(int64)
		if !ok {
			returnErrorJson(w, http.StatusUnauthorized, "Invalid user ID")
			return
		}

		settingName := "检索提示词"
		// if chatReq.CategoryId > 0 {
		// 	settingName = "rag-llm-" + strconv.FormatInt(chatReq.CategoryId, 10)
		// }
		logx.Debugf("AI session chat settingName=%s", settingName)
		//====================================================
		//get AI_FORMAT_STRING from database
		var AI_FORMAT_STRING []byte
		clientId, _ := r.Context().Value("clientId").(string)
		if strings.TrimSpace(clientId) == "" {
			returnErrorJson(w, http.StatusUnauthorized, "clientId不能为空，请重新登录")
			return
		}

		// 检查并设置默认配置
		getOrSetDefaultAIConfig := func(ctx context.Context, clientId, configName, defaultValue string) (string, error) {
			config, err := svcCtx.AiConfigModel.FindByNameAndCode(ctx, clientId, configName, "")
			if err == nil {
				if config != nil {
					return config.Value, nil
				}
			}

			// 配置不存在，创建默认配置
			newConfig := &model.AiConfig{
				ClientId:     clientId,
				DocumentCode: "",
				Name:         configName,
				Value:        defaultValue,
				CreatedBy:    "system",
			}

			_, err = svcCtx.AiConfigModel.Insert(ctx, newConfig)
			if err != nil {
				logx.Errorf("创建默认AI配置失败 [%s]: %v", configName, err)
				return defaultValue, err
			}

			logx.Infof("已为租户 %s 创建默认AI配置: %s", clientId, configName)
			return defaultValue, nil
		}

		// 获取检索提示词
		retrievalPrompt, err := getOrSetDefaultAIConfig(r.Context(), clientId, "检索提示词", "%s --- 根据以上参考资料，使用原文回答问题 \"%s\" ； 如果以上参考资料中没有问题的答案，就说明找不到这个问题的答案。")
		if err != nil {
			returnErrorJson(w, http.StatusInternalServerError, "读取AI配置 [检索提示词]失败: "+err.Error())
			return
		}
		AI_FORMAT_STRING = []byte(retrievalPrompt)

		// 获取角色提示词
		rolePrompt, err := getOrSetDefaultAIConfig(r.Context(), clientId, "角色提示词", "你是智能助手， 现在，请根据用户的具体问题提供专业、可靠、合规的协助。")
		if err != nil {
			logx.Errorf("获取角色提示词失败: %v", err)
			rolePrompt = "你是智能助手， 现在，请根据用户的具体问题提供专业、可靠、合规的协助。"
		}
		var AI_ROBOTPROMPT string
		AI_ROBOTPROMPT = rolePrompt

		// 获取问候词
		_, err = getOrSetDefaultAIConfig(r.Context(), clientId, "问候词", "你好，我是智能助手")
		if err != nil {
			logx.Errorf("获取问候词失败: %v", err)
		}

		//====================================================
		var ragContext string
		var fileinfos []string
		var fileSimilarities []string
		var rawDocs []rawDocMeta
		var fullTextMs int64
		var mainSearchMs int64
		var subSearchMs int64
		rawSearchUsed := false
		vectorSearchUsed := false

		orchestrator := aiLogic.NewRagRetrievalOrchestrator(r.Context(), svcCtx)
		ragResult, ragErr := orchestrator.Retrieve(aiLogic.RagRetrieveRequest{
			ClientId:        clientId,
			Message:         chatReq.Message,
			DocumentCode:    chatReq.DocumentCode,
			Tags:            chatReq.Tags,
			SkipRag:         chatReq.Skiprag,
			HasUploadedDocs: hasUploadedDocs,
		})
		if ragErr != nil {
			logx.Errorf("RAG orchestrator 执行失败: %v", ragErr)
		}
		if ragResult != nil {
			ragContext = ragResult.RagContext
			fileinfos = ragResult.FileInfos
			fileSimilarities = ragResult.FileSimilarities
			fullTextMs = ragResult.FullTextMs
			mainSearchMs = ragResult.MainSearchMs
			subSearchMs = ragResult.SubSearchMs
			rawSearchUsed = ragResult.RawSearchUsed
			vectorSearchUsed = ragResult.VectorSearchUsed
			for _, item := range ragResult.RawDocs {
				rawDocs = append(rawDocs, rawDocMeta{Id: item.Id, FileName: item.FileName})
			}
		}
		if ragContext != "" && len(AI_FORMAT_STRING) > 0 {
			chatReq.Message = fmt.Sprintf(string(AI_FORMAT_STRING), ragContext, chatReq.Message)
		}

		if chatReq.Test {
			//返回纯文本
			w.Header().Set("Content-Type", "text/plain")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(ragContext))

			return
		}

		tagsStr := strings.Join(chatReq.Tags, ",")

		if chatReq.Prompt == "" {
			chatReq.Prompt = "用中文回答用户"
		}

		// NO_THINK_SUFFIX := "/no_think"

		// Load AI config to get Temperature, TopK, TopP, MaxTokens
		aiConfig, err := knowsourceLogic.LoadAIConfig(clientId)
		if err != nil {
			logx.Errorf("加载AI配置失败: %v", err)
			// 使用默认值
			aiConfig = &types.LLMSettingData{
				Temperature: 0.7,
				TopK:        40,
				TopP:        0.9,
				MaxTokens:   4096,
			}
		}

		// Determine AI model: 请求指定 > LLM 配置中的 model（ai.yaml）> 默认 qwen3
		modelname := ""
		if chatReq.Model != "" {
			modelname = chatReq.Model
		} else {
			modelname = aiConfig.Model
		}

		if modelname == "" {
			modelname = "qwen3"
		}
		if modelname == "qwen3" {
			modelname = "Qwen3-0.6B"
		}
		if chatReq.Model == "deepseek" {
			modelname = "nezahatkorkmaz/deepseek-v3"
		}
		// 使用内存中的模型列表解析：若指定模型不在列表中，则选 qwen3 且不含 embedding 的
		modelname = utils.LLMModelStore.ResolveChatModel(modelname)

		// Get or create session from database
		session, err := getOrCreateSessionFromDB(r.Context(), svcCtx, chatReq.Session, userID, modelname,
			chatReq.CategoryId, chatReq.Keys, empCode, chatReq.DocumentCode, tagsStr)
		if err != nil {
			logx.Errorf("获取或创建会话失败: %v", err)
			returnErrorJson(w, http.StatusInternalServerError, "获取或创建会话失败"+err.Error())
			return
		}

		// Save original user message to database (before RAG processing)
		// originalUserMessage := chatReq.Message
		//TODO: originalUserMessage
		logx.Debugf("AI session chat userMessageLen=%d hasUploadedDocs=%v", len(originalUserMessage), hasUploadedDocs)

		userFiles := []string{}

		err = saveMessageToDB(r.Context(), svcCtx, session.SessionId, "user", originalUserMessage, modelname, "", "", userFiles)
		if err != nil {
			logx.Errorf("保存用户消息失败: %v", err)
		}

		// 同一 session 同一日志文件，按 clientid、session 与用户中文名命名，追加本次用户问题（写入 chatReq.Message，含 RAG 等处理后的内容）
		logPath := sessionLogPath(clientId, session.SessionId, userNameCn)

		// 写入找到的文件名称和相似度到日志
		if len(fileSimilarities) > 0 {
			filesStr := "找到的文件（概要检索）：" + strings.Join(fileSimilarities, ", ")
			appendSessionLog(logPath, "system", filesStr)
		}

		// 写入找到的文件名称到日志
		if len(fileinfos) > 0 {
			filesStr := "找到的文件：" + strings.Join(fileinfos, ", ")
			appendSessionLog(logPath, "system", filesStr)
		}

		appendSessionLog(logPath, "user", chatReq.Message)

		// Check if user wants to disable thinking

		// Prepare AI request with conversation history

		messages := []Message{}

		// Add /no_think suffix in Prompt if thinking is disabled
		// if !chatReq.Think && !strings.Contains(chatReq.Message, NO_THINK_SUFFIX) {
		// 	chatReq.Message = chatReq.Message + NO_THINK_SUFFIX
		// }

		// if !chatReq.Think && !strings.Contains(chatReq.Prompt, NO_THINK_SUFFIX) {
		// 	chatReq.Prompt = chatReq.Prompt + NO_THINK_SUFFIX
		// }

		// Add system message - use provided prompt or default Chinese response instruction

		if chatReq.Prompt != "" {
			messages = append(messages, Message{
				Role:    "system",
				Content: AI_ROBOTPROMPT + chatReq.Prompt,
			})
		} else {
			// Add default system message when no prompt is provided
			messages = append(messages, Message{
				Role:    "system",
				Content: AI_ROBOTPROMPT,
			})
		}

		// Load conversation history from database (last 20 messages to avoid token limit)
		historyMessages, err := svcCtx.AiMessagesModel.FindBySessionId(r.Context(), session.SessionId)
		if err != nil {
			logx.Errorf("获取会话历史失败: %v", err)
		} else {
			// Add last 20 messages to avoid token limit
			historyStart := 0
			if len(historyMessages) > 20 {
				historyStart = len(historyMessages) - 20
			}

			for i, msg := range historyMessages[historyStart:] {
				if msg.Context != "" {
					msg.Content = msg.Context + msg.Content
				}
				messages = append(messages, Message{
					Role:    msg.Role,
					Content: msg.Content,
				})
				logx.Debugf("历史消息 %d role=%s contentLen=%d", i, msg.Role, len(msg.Content))
			}
		}

		//update the last user message or add new user message
		// If model contains "qwen" (but is not qwen3.5), add /no_think prefix to the message when thinking is disabled
		userMessageContent := chatReq.Message

		lowerModel := strings.ToLower(modelname)
		if strings.Contains(lowerModel, "qwen") && !chatReq.Think && !strings.Contains(lowerModel, "qwen3.5") {
			if !strings.HasPrefix(userMessageContent, "/no_think ") {
				userMessageContent = "/no_think " + userMessageContent
				logx.Debugf("模型 %s 包含 qwen，已添加 /no_think 前缀", modelname)
			}
		}

		if len(messages) > 0 && messages[len(messages)-1].Role == "user" {
			messages[len(messages)-1].Content = userMessageContent
		} else {
			// Add new user message if last message is not user message
			messages = append(messages, Message{
				Role:    "user",
				Content: userMessageContent,
			})
		}

		aiReq := AIRequest{
			Model:       modelname,
			Messages:    messages,
			Stream:      true,
			Think:       chatReq.Think,
			Temperature: aiConfig.Temperature,
			TopK:        aiConfig.TopK,
			TopP:        aiConfig.TopP,
			MaxTokens:   aiConfig.MaxTokens,
		}

		// Convert request to JSON
		jsonData, err := json.Marshal(aiReq)
		if err != nil {
			logx.Errorf("转json失败: %v", err)
			returnErrorJson(w, http.StatusInternalServerError, "转json失败: "+err.Error())
			return
		}

		// Make request to AI API
		logx.Debugf("call ai stream request bodyLen=%d", len(jsonData))

		callStart := time.Now()
		var firstTokenMs int64
		var totalStreamMs int64
		var capturedResponse string
		var stats map[string]interface{}
		flusher, _ := w.(http.Flusher)

		// 若确认为 Ollama 配置，则使用 Ollama 原生 /api/chat 方式调用
		ollamaURL := utils.GetOllamaChatURL(&svcCtx.Config)
		if ollamaURL != "" {
			ollamaReq := struct {
				Model    string    `json:"model"`
				Messages []Message `json:"messages"`
				Think    bool      `json:"think"`
				Stream   bool      `json:"stream"`
			}{
				Model:    modelname,
				Messages: aiReq.Messages,
				Think:    chatReq.Think,
				Stream:   true,
			}
			ollamaBody, err := json.Marshal(ollamaReq)
			if err != nil {
				logx.Errorf("Ollama 请求 marshal 失败: %v", err)
				returnErrorJson(w, http.StatusInternalServerError, "Ollama 请求 marshal 失败: "+err.Error())
				return
			}
			logx.Infof("使用 Ollama 原生 API: %s bodyLen=%d streamTimeout=%s", ollamaURL, len(ollamaBody), llmHTTPStreamTimeout)
			llmStreamCtx, streamCancel := context.WithTimeout(r.Context(), llmHTTPStreamTimeout)
			defer streamCancel()
			httpReq, err := http.NewRequestWithContext(llmStreamCtx, "POST", ollamaURL, bytes.NewBuffer(ollamaBody))
			if err != nil {
				logx.Errorf("创建 Ollama 请求失败: %v", err)
				returnErrorJson(w, http.StatusInternalServerError, "创建 Ollama 请求失败: "+err.Error())
				return
			}
			httpReq.Header.Set("Content-Type", "application/json")
			resp, err := newLLMHTTPClient(false).Do(httpReq)
			if err != nil {
				logx.Errorf("联系 Ollama API 失败: %v", err)
				returnErrorJson(w, http.StatusBadGateway, "联系 Ollama API 失败: "+err.Error())
				return
			}
			defer resp.Body.Close()
			if resp.StatusCode != http.StatusOK {
				body, _ := io.ReadAll(resp.Body)
				logx.Errorf("Ollama API 返回 %d: %s", resp.StatusCode, string(body))
				returnErrorJson(w, resp.StatusCode, "Ollama API 返回 "+strconv.Itoa(resp.StatusCode)+": "+string(body))
				return
			}

			w.Header().Set("Content-Type", "text/event-stream")
			w.Header().Set("llmbackend", "ollama")
			w.Header().Set("SessionUuid", session.SessionUuid)
			w.WriteHeader(http.StatusOK)

			stats = map[string]interface{}{
				"fullTextMs":    fullTextMs,
				"mainSearchMs":  mainSearchMs,
				"subSearchMs":   subSearchMs,
				"modelName":     modelname,
				"rawSearchUsed": rawSearchUsed,
				"vectorUsed":    vectorSearchUsed,
			}

			_, cap := writeStreamMetaAndCopy(w, flusher, fileinfos, rawDocs, resp.Body, stats, callStart, &firstTokenMs, &totalStreamMs)
			capturedResponse = cap
		} else {
			// 统一按 OpenAI 兼容方式调用 Llm.CompletionUrl
			baseURL, _, completionApiKey := utils.ResolveCompletionRuntime(&svcCtx.Config, clientId)
			apiUrl := ""
			if baseURL != "" {
				apiUrl = strings.TrimSuffix(baseURL, "/") + "/v1/chat/completions"
			}
			var requestBody []byte

			if apiUrl == "" {
				logx.Errorf("未配置 LLM 地址（Llm.CompletionUrl）")
				returnErrorJson(w, http.StatusServiceUnavailable, "未配置 LLM 地址（Llm.CompletionUrl）")
				return
			}

			openAIReq := OpenAIRequest{
				Model:       modelname,
				Messages:    aiReq.Messages,
				Stream:      aiReq.Stream,
				Temperature: aiReq.Temperature,
				TopP:        aiReq.TopP,
				MaxTokens:   aiReq.MaxTokens,
			}
			if !chatReq.Think {
				lowerModel := strings.ToLower(modelname)
				if strings.Contains(lowerModel, "qwen3.5") {
					openAIReq.ExtraBody = map[string]interface{}{
						"enable_thinking": false,
						"reasoning":       false,
					}
				}
			}
			requestBody, err = json.Marshal(openAIReq)
			if err != nil {
				logx.Errorf("转json失败: %v", err)
				returnErrorJson(w, http.StatusInternalServerError, "转json失败: "+err.Error())
				return
			}
			logx.Infof("使用 OpenAI 兼容 API: %s bodyLen=%d streamTimeout=%s", apiUrl, len(requestBody), llmHTTPStreamTimeout)
			llmStreamCtx, streamCancel := context.WithTimeout(r.Context(), llmHTTPStreamTimeout)
			defer streamCancel()
			httpReq, err := http.NewRequestWithContext(llmStreamCtx, "POST", apiUrl, bytes.NewBuffer(requestBody))
			if err != nil {
				logx.Errorf("创建请求失败: %v", err)
				returnErrorJson(w, http.StatusInternalServerError, "创建请求失败: "+err.Error())
				return
			}
			httpReq.Header = r.Header.Clone()
			httpReq.Header.Set("Content-Type", "application/json")
			if strings.TrimSpace(completionApiKey) != "" {
				httpReq.Header.Set("Authorization", "Bearer "+strings.TrimSpace(completionApiKey))
			}

			resp, err := newLLMHTTPClient(false).Do(httpReq)
			if err != nil {
				errinfo := fmt.Sprintf("联系AI API失败: %v\napiUrl: %s", err, apiUrl)
				logx.Errorf("%s", errinfo)
				returnErrorJson(w, http.StatusBadGateway, errinfo)
				return
			}
			defer resp.Body.Close()

			for k, v := range resp.Header {
				if strings.EqualFold(k, "Content-Length") {
					continue
				}
				for _, vv := range v {
					w.Header().Add(k, vv)
				}
			}
			w.Header().Set("SessionUuid", session.SessionUuid)
			w.WriteHeader(resp.StatusCode)

			stats = map[string]interface{}{
				"fullTextMs":    fullTextMs,
				"mainSearchMs":  mainSearchMs,
				"subSearchMs":   subSearchMs,
				"modelName":     modelname,
				"rawSearchUsed": rawSearchUsed,
				"vectorUsed":    vectorSearchUsed,
			}

			_, capturedResponse = writeStreamMetaAndCopy(w, flusher, fileinfos, rawDocs, resp.Body, stats, callStart, &firstTokenMs, &totalStreamMs)
		}

		// Parse the captured response to extract AI content and model
		parsed := parseStreamResponse(capturedResponse)
		aiResponse := parsed.Content
		aiThinking := parsed.Reason
		modelForStats := parsed.Model
		if modelForStats == "" {
			modelForStats = modelname
		}
		logx.Infof("stream 完成 respLen=%d model=%s fileCount=%d", len(aiResponse), modelForStats, len(fileinfos))
		logx.Debugf("stream 完成 fileinfos=%v", fileinfos)

		// 在写入 assistant 消息前，用“答案 vs 文件名”做一次 rerank 归因，尽量定位本次回答引用了哪个文档
		// 命中阈值：0.8（可后续做成配置项）
		attributedFiles := []string{}
		if strings.TrimSpace(aiResponse) != "" && len(fileinfos) > 0 {
			qc := svcCtx.QdrantTools
			if qc == nil && svcCtx.QdrantClient != nil {
				qc = utils.NewQdrantToolsWithClient(svcCtx.QdrantClient)
			}
			if qc != nil {
				rerankURL, rerankType, rerankApiKey, _ := utils.ResolveRerankRuntime(&svcCtx.Config, clientId)
				matched, rerankResults, rerr := qc.FindDocFilenamesByRerankWithAPIKey(
					r.Context(),
					RemoveThinkPrefix(aiResponse),
					fileinfos,
					rerankURL,
					rerankApiKey,
					rerankType,
					0.8,
				)
				if rerr != nil {
					logx.Errorf("答案-文件名 rerank 归因失败: %v", rerr)
				} else {
					attributedFiles = matched
					if len(attributedFiles) > 0 {
						appendSessionLog(logPath, "system", "答案引用文件(阈值0.8)："+strings.Join(attributedFiles, ", "))
					} else if len(rerankResults) > 0 {
						// 无命中也写入 top1，便于排查（不影响 DB 存储）
						top := rerankResults[0]
						if top.Index >= 0 && top.Index < len(fileinfos) {
							appendSessionLog(logPath, "system", fmt.Sprintf("答案引用文件未达阈值(0.8)，top1=%s score=%.4f", fileinfos[top.Index], top.RelevanceScore))
						}
					}
				}
			}
		}

		// 将归因后的文件名信息输出给前端（流结束后的 meta 事件）
		if len(attributedFiles) > 0 {
			fileinfos = attributedFiles
		}
		_ = writeStreamMeta(w, flusher, fileinfos, rawDocs, stats, attributedFiles)

		// 写入 ai_call_stats 统计表（使用 stream 返回的 model 若存在）
		userIDStr := empCode
		costMs := uint64(time.Since(callStart).Milliseconds())
		questionChars := uint64(len([]rune(originalUserMessage)))
		outputChars := uint64(len([]rune(aiResponse)))
		if _, errStats := svcCtx.AiCallStatsModel.Insert(r.Context(), &model.AiCallStats{
			UserId:            userIDStr,
			CallTime:          time.Now(),
			CostMs:            costMs,
			QuestionCharCount: questionChars,
			OutputCharCount:   outputChars,
			ModelName:         modelForStats,
			CallStatus:        1,
		}); errStats != nil {
			logx.Errorf("写入 ai_call_stats 失败: %v", errStats)
		}

		// Save AI response to database
		finalFiles := fileinfos
		if len(attributedFiles) > 0 {
			finalFiles = attributedFiles
		}
		err = saveMessageToDB(r.Context(), svcCtx, session.SessionId, "assistant", RemoveThinkPrefix(aiResponse), modelForStats, ragContext, aiThinking, finalFiles)
		if err != nil {
			logx.Errorf("保存AI回复失败: %v", err)
		}

		// 同一 session 同一日志文件，追加本次 AI 回复
		appendSessionLog(logPath, "assistant", RemoveThinkPrefix(aiResponse))
	}
}

// // rowDatatoMD converts filesWithRows to markdown format
// func rowDatatoMD(filesWithRows []*knowdata.KnowledgeDataFileWithRows, keys string) []string {
// 	var mdStrings []string

// 	ii := 1

// 	for _, fileWithRows := range filesWithRows {

// 		keyspath := strings.ReplaceAll(keys, ",", "/")
// 		if !strings.Contains(fileWithRows.FilePath, keyspath) {
// 			continue
// 		}
// 		// Parse JSON rowData

// 		header, row, err := hdLogic.FormatRowDataAsArray(fileWithRows.RowHeader, fileWithRows.RowData)
// 		if err != nil {
// 			logx.Errorf("解析rowData JSON失败: %v", err)
// 			continue
// 		}
// 		rowDataMap := hdLogic.ConvertRowDataAsMap(header, row)

// 		// var rowDataMap map[string]interface{}
// 		// if err := json.Unmarshal([]byte(fileWithRows.RowData), &rowDataMap); err != nil {
// 		// 	logx.Errorf("解析rowData JSON失败: %v", err)
// 		// 	continue
// 		// }

// 		var mdContent strings.Builder

// 		mdContent.WriteString(fmt.Sprintf("# 精准匹配文档 %d\n", ii))
// 		ii++

// 		// Iterate through the map and create markdown sections
// 		for key, value := range rowDataMap {
// 			if value == "" || fmt.Sprintf("%v", value) == "" {
// 				continue
// 			}

// 			// Clean up field name
// 			fieldName := strings.ReplaceAll(key, "\n", "")
// 			if strings.TrimSpace(fieldName) == "" {
// 				fieldName = "内容"
// 			}

// 			mdContent.WriteString(fmt.Sprintf("### %s\n%v\n", fieldName, value))
// 		}

// 		mdStrings = append(mdStrings, mdContent.String())
// 	}

// 	return mdStrings
// }

var thinkPrefixRE = regexp.MustCompile(`^<think>[\s]*</think>`)
var thinkBlockRE = regexp.MustCompile(`(?is)<think>(.*?)</think>`)

func RemoveThinkPrefix(msg string) string {
	// 正则表达式解释：
	// <think> 匹配开头标签
	// [\s]*   匹配任意数量的空白字符（空格、换行、制表符等）
	// </think> 匹配结尾标签
	// ^ 锚定开头，确保只匹配前缀的该标签块（避免匹配中间的）
	return thinkPrefixRE.ReplaceAllString(msg, "")
}

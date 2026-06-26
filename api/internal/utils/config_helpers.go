package utils

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"knowsource/api/internal/config"

	"github.com/zeromicro/go-zero/core/logx"
)

// embeddingModelOverride 来自 ai.yaml 的 embeddingModel，由 LoadLLMSetting/SaveLLMSetting 设置，按 clientId 存储
var (
	embeddingModelOverrides map[string]string
	embeddingModelMu        sync.RWMutex
)

func init() {
	embeddingModelOverrides = make(map[string]string)
}

// SetEmbeddingModelOverride 设置 ai.yaml 中的 embedding 模型（LLM 设置保存/加载时调用）
func SetEmbeddingModelOverride(clientId, model string) {
	embeddingModelMu.Lock()
	defer embeddingModelMu.Unlock()
	embeddingModelOverrides[clientId] = model
}

// GetEmbeddingModelOverride 获取 ai.yaml 中配置的 embedding 模型
func GetEmbeddingModelOverride(clientId string) string {
	embeddingModelMu.RLock()
	defer embeddingModelMu.RUnlock()
	return embeddingModelOverrides[clientId]
}

// GetEmbeddingModelOverrideWithDefault 为保持向后兼容，支持不传递 clientId
func GetEmbeddingModelOverrideWithDefault() string {
	return ""
}

// versionResponse 用于解析 /api/version 或 /version 的 JSON 返回，如 {"version":"0.1"}
type versionResponse struct {
	Version string `json:"version"`
}

// modelsResponse 用于解析 /v1/models 的 data 数组，检查 owned_by
type modelsResponse struct {
	Data []struct {
		OwnedBy string `json:"owned_by"`
	} `json:"data"`
}

// tryModelsForLlamaCpp 当 /api/version 和 /version 都失败时，尝试 /v1/models；若存在 owned_by=="llamacpp" 则返回 llamacpp 类型
func tryModelsForLlamaCpp(client *http.Client, base string) (ctype, version string, err error) {
	u := base + "/v1/models"
	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, u, nil)
	if err != nil {
		return "", "", fmt.Errorf("neither /api/version (ollama) nor /version (vllm) reachable, and /v1/models request failed: %w", err)
	}
	resp, err := client.Do(req)
	if err != nil {
		return "", "", fmt.Errorf("neither /api/version (ollama) nor /version (vllm) reachable, and /v1/models unreachable: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", "", fmt.Errorf("neither /api/version nor /version reachable, /v1/models returned HTTP %d", resp.StatusCode)
	}
	var m modelsResponse
	if err := json.NewDecoder(resp.Body).Decode(&m); err != nil {
		return "", "", fmt.Errorf("neither /api/version nor /version reachable, /v1/models response invalid: %w", err)
	}
	for _, item := range m.Data {
		if item.OwnedBy == "llamacpp" {
			return "llamacpp", "", nil
		}
	}
	return "", "", fmt.Errorf("neither /api/version (ollama) nor /version (vllm) reachable, and /v1/models has no owned_by=llamacpp")
}

// GetCompletionTypeAndVersion 探测 base URL，返回 type（ollama/vllm/llamacpp）和 version。
// - Ollama: 可访问 /api/version，返回 {"version":"..."}
// - vLLM: 可访问 /version，返回 {"version":"..."}
// - llama.cpp: 若上述两者都失败，请求 /v1/models，若存在 owned_by=="llamacpp" 则类型为 llamacpp，版本未知
func GetCompletionTypeAndVersion(baseURL string) (ctype, version string, err error) {
	base := strings.TrimSuffix(strings.TrimSpace(baseURL), "/")
	if base == "" {
		return "", "", fmt.Errorf("CompletionUrl is empty")
	}
	client := &http.Client{Timeout: 5 * time.Second}
	ctx := context.Background()

	// 1. 先试 /api/version（Ollama）
	uOllama := base + "/api/version"
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, uOllama, nil)
	if err != nil {
		return "", "", err
	}
	resp, err := client.Do(req)
	if err == nil {
		if resp.StatusCode == http.StatusOK {
			var v versionResponse
			if errDec := json.NewDecoder(resp.Body).Decode(&v); errDec == nil && v.Version != "" {
				resp.Body.Close()
				return "ollama", v.Version, nil
			}
		}
		if resp != nil && resp.Body != nil {
			resp.Body.Close()
		}
	}

	// 2. 再试 /version（vLLM）
	uVllm := base + "/version"
	req, err = http.NewRequestWithContext(ctx, http.MethodGet, uVllm, nil)
	if err != nil {
		return "", "", err
	}
	resp, err = client.Do(req)
	if err != nil {
		return tryModelsForLlamaCpp(client, base)
	}
	if resp != nil {
		defer resp.Body.Close()
	}
	if resp == nil || resp.StatusCode != http.StatusOK {
		return tryModelsForLlamaCpp(client, base)
	}
	var v versionResponse
	if err := json.NewDecoder(resp.Body).Decode(&v); err != nil || v.Version == "" {
		return tryModelsForLlamaCpp(client, base)
	}
	return "vllm", v.Version, nil
}

// DetectCompletionType 通过探测 base URL 确定是 ollama 还是 vllm（用于启动时配置）
func DetectCompletionType(baseURL string) (string, error) {
	t, version, err := GetCompletionTypeAndVersion(baseURL)
	logx.Infof("version: %s", version)
	return t, err
}

// EnsureCompletionType 读取配置后，根据 CompletionUrl 探测并设置 CompletionType（vllm/ollama/llamacpp）
func EnsureCompletionType(cfg *config.Config) {
	if cfg.Llm.CompletionUrl == "" {
		return
	}
	base := strings.TrimSuffix(cfg.Llm.CompletionUrl, "/")
	detected, err := DetectCompletionType(base)
	if err != nil {
		logx.Errorf("检测 Llm.CompletionType 失败（保留原配置 %q）: %v", cfg.Llm.CompletionType, err)
		return
	}
	cfg.Llm.CompletionType = detected
	logx.Infof("Llm.CompletionType 已自动检测为: %s", detected)
}

// GetCompletionURL 返回 OpenAI 兼容的 /v1/chat/completions 完整 URL
func GetCompletionURL(cfg *config.Config) string {
	base := strings.TrimSpace(cfg.Llm.CompletionUrl)
	if base != "" {
		return strings.TrimSuffix(base, "/") + "/v1/chat/completions"
	}
	return ""
}

// GetChatBaseForModels 返回用于 GET /v1/models 或 /api/tags 的 base URL
func GetChatBaseForModels(cfg *config.Config) string {
	if cfg.Llm.CompletionUrl != "" {
		return strings.TrimSuffix(cfg.Llm.CompletionUrl, "/")
	}
	return ""
}

// GetOllamaChatURL 当 CompletionType 为 ollama 时返回 /api/chat 完整 URL，否则返回空
func GetOllamaChatURL(cfg *config.Config) string {
	if strings.TrimSpace(strings.ToLower(cfg.Llm.CompletionType)) != "ollama" {
		return ""
	}
	base := GetChatBaseForModels(cfg)
	if base == "" {
		return ""
	}
	return base + "/api/chat"
}

// ApplyEmbeddingConfig 根据配置设置 QdrantTools 的 Embedding
// Rag 时按 EmbeddingsType 选择 vllm 或 ollama
func ApplyEmbeddingConfig(cfg *config.Config, qc *QdrantTools) error {
	return ApplyEmbeddingConfigForClient(cfg, qc, "")
}

func ApplyEmbeddingConfigForClient(cfg *config.Config, qc *QdrantTools, clientId string) error {
	base, t, apiKey, resolved := ResolveEmbeddingRuntime(cfg, clientId)
	if base != "" {
		if resolved != "" {
			qc.SetEmbeddingModel(resolved)
		}
		if t == "vllm" {
			qc.SetVllmEmbedding(base, resolved)
			qc.SetVllmEmbeddingAPIKey(apiKey)
			return nil
		}
		// ollama：Embedding 接口在 /api/embed，base 如 http://host:11434
		ollamaBase := strings.TrimSuffix(base, "/api/chat")
		ollamaBase = strings.TrimSuffix(ollamaBase, "/api/embed")
		ollamaBase = strings.TrimSuffix(ollamaBase, "/")
		qc.SetEmbeddingAPI(ollamaBase)
		qc.SetEmbeddingAPIKey(apiKey)
		return nil
	}
	return fmt.Errorf("未配置向量化地址（Rag.EmbeddingsUrl）")
}

// NewQdrantToolsFromConfig 使用系统配置初始化 QdrantTools（仅连接能力，不注入 embedding）
func NewQdrantToolsFromConfig(cfg *config.Config) (*QdrantTools, error) {
	if cfg == nil {
		return nil, fmt.Errorf("配置为空")
	}
	if cfg.Qdrant.Host == "" || cfg.Qdrant.Port <= 0 {
		return nil, fmt.Errorf("未配置 Qdrant 地址")
	}
	return NewQdrantTools(cfg.Qdrant.Host, cfg.Qdrant.Port), nil
}

// NewQdrantToolsWithEmbedding 使用系统配置初始化 QdrantTools，并注入 embedding 配置
func NewQdrantToolsWithEmbedding(cfg *config.Config) (*QdrantTools, error) {
	return NewQdrantToolsWithEmbeddingForClient(cfg, "")
}

func NewQdrantToolsWithEmbeddingForClient(cfg *config.Config, clientId string) (*QdrantTools, error) {
	qc, err := NewQdrantToolsFromConfig(cfg)
	if err != nil {
		return nil, err
	}
	if err := ApplyEmbeddingConfigForClient(cfg, qc, clientId); err != nil {
		return nil, err
	}
	return qc, nil
}

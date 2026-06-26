package utils

import (
	"os"
	"path/filepath"
	"strings"
	"sync"

	"knowsource/api/internal/config"

	"github.com/spf13/viper"
)

type TenantLLMOverride struct {
	Model          string
	EmbeddingModel string
	RerankerModel  string

	CompletionURL     string
	CompletionType    string
	CompletionAPIKey  string
	EmbeddingsURL     string
	EmbeddingsType    string
	EmbeddingsAPIKey  string
	RerankerURL       string
	RerankerType      string
	RerankerAPIKey    string
	RagEmbeddingTopK  int64
	RagRerankTopK     int64
}

var (
	tenantLLMOverrideMu   sync.RWMutex
	tenantLLMOverrideData = map[string]TenantLLMOverride{}
)

func SetTenantLLMOverride(clientId string, cfg TenantLLMOverride) {
	clientId = strings.TrimSpace(clientId)
	if clientId == "" {
		return
	}
	tenantLLMOverrideMu.Lock()
	defer tenantLLMOverrideMu.Unlock()
	tenantLLMOverrideData[clientId] = cfg
}

func getTenantLLMOverride(clientId string) TenantLLMOverride {
	clientId = strings.TrimSpace(clientId)
	if clientId == "" {
		return TenantLLMOverride{}
	}
	tenantLLMOverrideMu.RLock()
	v, ok := tenantLLMOverrideData[clientId]
	tenantLLMOverrideMu.RUnlock()
	if ok {
		return v
	}
	loaded := loadTenantLLMOverrideFromFile(clientId)
	tenantLLMOverrideMu.Lock()
	tenantLLMOverrideData[clientId] = loaded
	tenantLLMOverrideMu.Unlock()
	return loaded
}

func loadTenantLLMOverrideFromFile(clientId string) TenantLLMOverride {
	filePath, err := getTenantAIConfigFilePath(clientId)
	if err != nil {
		return TenantLLMOverride{}
	}
	if _, statErr := os.Stat(filePath); statErr != nil {
		return TenantLLMOverride{}
	}
	v := viper.New()
	v.SetConfigFile(filePath)
	v.SetConfigType("yaml")
	if readErr := v.ReadInConfig(); readErr != nil {
		return TenantLLMOverride{}
	}
	return TenantLLMOverride{
		Model:             strings.TrimSpace(v.GetString("model")),
		EmbeddingModel:    strings.TrimSpace(v.GetString("embeddingModel")),
		RerankerModel:     strings.TrimSpace(v.GetString("rerankerModel")),
		CompletionURL:     strings.TrimSpace(v.GetString("completionUrl")),
		CompletionType:    strings.TrimSpace(v.GetString("completionType")),
		CompletionAPIKey:  strings.TrimSpace(v.GetString("completionApiKey")),
		EmbeddingsURL:     strings.TrimSpace(v.GetString("embeddingsUrl")),
		EmbeddingsType:    strings.TrimSpace(v.GetString("embeddingsType")),
		EmbeddingsAPIKey:  strings.TrimSpace(v.GetString("embeddingsApiKey")),
		RerankerURL:       strings.TrimSpace(v.GetString("rerankerUrl")),
		RerankerType:      strings.TrimSpace(v.GetString("rerankerType")),
		RerankerAPIKey:    strings.TrimSpace(v.GetString("rerankerApiKey")),
		RagEmbeddingTopK:  v.GetInt64("ragEmbeddingTopK"),
		RagRerankTopK:     v.GetInt64("ragRerankTopK"),
	}
}

func getTenantAIConfigFilePath(clientId string) (string, error) {
	execPath, err := os.Executable()
	if err != nil {
		return "", err
	}
	execDir := filepath.Dir(execPath)
	return filepath.Join(execDir, "ai_"+clientId+".yaml"), nil
}

func ResolveCompletionRuntime(cfg *config.Config, clientId string) (baseURL, ctype, apiKey string) {
	baseURL = strings.TrimSpace(cfg.Llm.CompletionUrl)
	ctype = strings.TrimSpace(cfg.Llm.CompletionType)
	over := getTenantLLMOverride(clientId)
	if over.CompletionURL != "" {
		baseURL = over.CompletionURL
	}
	if over.CompletionType != "" {
		ctype = over.CompletionType
	}
	apiKey = over.CompletionAPIKey
	return strings.TrimSuffix(baseURL, "/"), strings.ToLower(strings.TrimSpace(ctype)), strings.TrimSpace(apiKey)
}

func ResolveEmbeddingRuntime(cfg *config.Config, clientId string) (baseURL, etype, apiKey, model string) {
	baseURL = strings.TrimSpace(cfg.Rag.EmbeddingsUrl)
	etype = strings.TrimSpace(cfg.Rag.EmbeddingsType)
	over := getTenantLLMOverride(clientId)
	if over.EmbeddingsURL != "" {
		baseURL = over.EmbeddingsURL
	}
	if over.EmbeddingsType != "" {
		etype = over.EmbeddingsType
	}
	apiKey = over.EmbeddingsAPIKey
	model = strings.TrimSpace(over.EmbeddingModel)
	if model == "" {
		model = strings.TrimSpace(GetEmbeddingModelOverride(clientId))
	}
	if model == "" {
		model = LLMModelStore.ResolveEmbeddingModel(DefaultEmbeddingModel)
	}
	return strings.TrimSuffix(baseURL, "/"), strings.ToLower(strings.TrimSpace(etype)), strings.TrimSpace(apiKey), model
}

func ResolveRerankRuntime(cfg *config.Config, clientId string) (baseURL, rtype, apiKey, model string) {
	baseURL = strings.TrimSpace(cfg.Rag.RerankerUrl)
	rtype = strings.TrimSpace(cfg.Rag.RerankerType)
	over := getTenantLLMOverride(clientId)
	if over.RerankerURL != "" {
		baseURL = over.RerankerURL
	}
	if over.RerankerType != "" {
		rtype = over.RerankerType
	}
	apiKey = over.RerankerAPIKey
	model = strings.TrimSpace(over.RerankerModel)
	if model == "" {
		model = LLMModelStore.ResolveRerankModel("")
	}
	return strings.TrimSuffix(baseURL, "/"), strings.ToLower(strings.TrimSpace(rtype)), strings.TrimSpace(apiKey), model
}

func ResolveChatModel(clientId, fallback string) string {
	over := getTenantLLMOverride(clientId)
	if strings.TrimSpace(over.Model) != "" {
		return strings.TrimSpace(over.Model)
	}
	if strings.TrimSpace(fallback) != "" {
		return strings.TrimSpace(fallback)
	}
	return LLMModelStore.ResolveChatModel("")
}

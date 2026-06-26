package knowsource

import (
	"context"
	"os"
	"path/filepath"
	"strings"

	"knowsource/api/internal/config"
	"knowsource/api/internal/svc"
	"knowsource/api/internal/types"
	"knowsource/api/internal/utils"
	"knowsource/common/response"

	"github.com/spf13/viper"
	"github.com/zeromicro/go-zero/core/logx"
)

type LoadLLMSettingLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 获取设置
func NewLoadLLMSettingLogic(ctx context.Context, svcCtx *svc.ServiceContext) *LoadLLMSettingLogic {
	return &LoadLLMSettingLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// getAIConfigFilePath 获取 ai.yaml 配置文件路径，根据 clientId 生成文件名
func getAIConfigFilePath(clientId string) (string, error) {
	execPath, err := os.Executable()
	if err != nil {
		return "", err
	}
	execDir := filepath.Dir(execPath)
	confpath := filepath.Join(execDir, "ai_"+clientId+".yaml")
	return confpath, nil
}

// BuildSystemLLMSettingDefaults 从主配置 knowsource.yaml（Llm/Rag）与内置缺省项生成系统默认 LLM 设置
func BuildSystemLLMSettingDefaults(cfg *config.Config) *types.LLMSettingData {
	if cfg == nil {
		cfg = &config.Config{}
	}
	return &types.LLMSettingData{
		MaxTokens:                      16384,
		Model:                          utils.LLMModelStore.ResolveChatModel(""),
		EmbeddingModel:                 utils.LLMModelStore.ResolveEmbeddingModel(""),
		CompletionUrl:                  strings.TrimSpace(cfg.Llm.CompletionUrl),
		CompletionType:                   strings.TrimSpace(cfg.Llm.CompletionType),
		EmbeddingsUrl:                  strings.TrimSpace(cfg.Rag.EmbeddingsUrl),
		EmbeddingsType:                 strings.TrimSpace(cfg.Rag.EmbeddingsType),
		RerankerUrl:                    strings.TrimSpace(cfg.Rag.RerankerUrl),
		RerankerType:                   strings.TrimSpace(cfg.Rag.RerankerType),
		RerankerModel:                  utils.LLMModelStore.ResolveRerankModel(""),
		Temperature:                    0.7,
		TopK:                           40,
		TopP:                           0.9,
		RepeatPenalty:                  1.1,
		RagEmbeddingTopK:               10,
		RagSimilarityThreshold:           0.3,
		RagSummarySimilarityThreshold:  0.3,
		RagRerankTopK:                  5,
		RagRerankScoreThreshold:          0.3,
		RagSummaryRerankScoreThreshold: 0.3,
	}
}

// LoadAIConfig 使用 viper 加载配置文件，如果不存在则创建默认配置，根据 clientId 读取对应文件
func LoadAIConfig(clientId string) (*types.LLMSettingData, error) {
	filePath, err := getAIConfigFilePath(clientId)
	if err != nil {
		return nil, err
	}

	// 检查文件是否存在
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		// 文件不存在，创建默认配置
		// 对话模型为空，读取第一个不带 embedding 的模型
		defaultModel := utils.LLMModelStore.ResolveChatModel("")
		// Embedding 模型为空，读取第一个带 embedding 的模型
		defaultEmbeddingModel := utils.LLMModelStore.ResolveEmbeddingModel("")
		defaultConfig := &types.LLMSettingData{
			MaxTokens:              16384,
			Model:                  defaultModel,
			EmbeddingModel:         defaultEmbeddingModel,
			CompletionUrl:          "",
			CompletionType:         "",
			EmbeddingsUrl:          "",
			EmbeddingsType:         "",
			RerankerUrl:            "",
			RerankerType:           "",
			RerankerModel:          utils.LLMModelStore.ResolveRerankModel(""),
			Temperature:            0.7,
			TopK:                   40,
			TopP:                   0.9,
			RepeatPenalty:          1.1,
			RagEmbeddingTopK:       10,
			RagSimilarityThreshold: 0.3,
			// 概要检索阈值：默认设置为 0.3
			RagSummarySimilarityThreshold: 0.3,
			RagRerankTopK:                 5,
			RagRerankScoreThreshold:       0.3,
			// 概要检索重排分值阈值：默认设置为 0.3
			RagSummaryRerankScoreThreshold: 0.3,
		}

		// 使用 viper 保存默认配置
		v := viper.New()
		v.SetConfigFile(filePath)
		v.SetConfigType("yaml")
		v.Set("maxSize", defaultConfig.MaxTokens)
		v.Set("model", defaultConfig.Model)
		v.Set("embeddingModel", defaultConfig.EmbeddingModel)
		v.Set("completionUrl", defaultConfig.CompletionUrl)
		v.Set("completionApiKey", defaultConfig.CompletionApiKey)
		v.Set("completionType", defaultConfig.CompletionType)
		v.Set("embeddingsUrl", defaultConfig.EmbeddingsUrl)
		v.Set("embeddingsApiKey", defaultConfig.EmbeddingsApiKey)
		v.Set("embeddingsType", defaultConfig.EmbeddingsType)
		v.Set("rerankerUrl", defaultConfig.RerankerUrl)
		v.Set("rerankerApiKey", defaultConfig.RerankerApiKey)
		v.Set("rerankerType", defaultConfig.RerankerType)
		v.Set("rerankerModel", defaultConfig.RerankerModel)
		v.Set("temperature", defaultConfig.Temperature)
		v.Set("topK", defaultConfig.TopK)
		v.Set("topP", defaultConfig.TopP)
		v.Set("repeatPenalty", defaultConfig.RepeatPenalty)
		v.Set("ragEmbeddingTopK", defaultConfig.RagEmbeddingTopK)
		v.Set("ragSimilarityThreshold", defaultConfig.RagSimilarityThreshold)
		v.Set("ragSummarySimilarityThreshold", defaultConfig.RagSummarySimilarityThreshold)
		v.Set("ragRerankTopK", defaultConfig.RagRerankTopK)
		v.Set("ragRerankScoreThreshold", defaultConfig.RagRerankScoreThreshold)
		v.Set("ragSummaryRerankScoreThreshold", defaultConfig.RagSummaryRerankScoreThreshold)

		if err := v.WriteConfigAs(filePath); err != nil {
			return nil, err
		}

		logx.Infof("Created default ai_%s.yaml config file at: %s", clientId, filePath)
		utils.SetEmbeddingModelOverride(clientId, defaultConfig.EmbeddingModel)
		utils.SetTenantLLMOverride(clientId, utils.TenantLLMOverride{
			Model:            defaultConfig.Model,
			EmbeddingModel:   defaultConfig.EmbeddingModel,
			RerankerModel:    defaultConfig.RerankerModel,
			CompletionURL:    defaultConfig.CompletionUrl,
			CompletionType:   defaultConfig.CompletionType,
			CompletionAPIKey: defaultConfig.CompletionApiKey,
			EmbeddingsURL:    defaultConfig.EmbeddingsUrl,
			EmbeddingsType:   defaultConfig.EmbeddingsType,
			EmbeddingsAPIKey: defaultConfig.EmbeddingsApiKey,
			RerankerURL:      defaultConfig.RerankerUrl,
			RerankerType:     defaultConfig.RerankerType,
			RerankerAPIKey:   defaultConfig.RerankerApiKey,
		})
		return defaultConfig, nil
	}

	// 文件存在，使用 viper 读取配置
	v := viper.New()
	v.SetConfigFile(filePath)
	v.SetConfigType("yaml")

	if err := v.ReadInConfig(); err != nil {
		return nil, err
	}

	config := &types.LLMSettingData{
		MaxTokens:              v.GetInt64("maxSize"),
		Model:                  v.GetString("model"),
		EmbeddingModel:         v.GetString("embeddingModel"),
		CompletionUrl:          v.GetString("completionUrl"),
		CompletionApiKey:       v.GetString("completionApiKey"),
		CompletionType:         v.GetString("completionType"),
		EmbeddingsUrl:          v.GetString("embeddingsUrl"),
		EmbeddingsApiKey:       v.GetString("embeddingsApiKey"),
		EmbeddingsType:         v.GetString("embeddingsType"),
		RerankerUrl:            v.GetString("rerankerUrl"),
		RerankerApiKey:         v.GetString("rerankerApiKey"),
		RerankerType:           v.GetString("rerankerType"),
		RerankerModel:          v.GetString("rerankerModel"),
		Temperature:            v.GetFloat64("temperature"),
		TopK:                   v.GetInt64("topK"),
		TopP:                   v.GetFloat64("topP"),
		RepeatPenalty:          v.GetFloat64("repeatPenalty"),
		RagEmbeddingTopK:       v.GetInt64("ragEmbeddingTopK"),
		RagSimilarityThreshold: v.GetFloat64("ragSimilarityThreshold"),
		// ragSummarySimilarityThreshold 兼容：未配置时沿用 ragSimilarityThreshold
		RagSummarySimilarityThreshold: func() float64 {
			if v.IsSet("ragSummarySimilarityThreshold") {
				return v.GetFloat64("ragSummarySimilarityThreshold")
			}
			return v.GetFloat64("ragSimilarityThreshold")
		}(),
		RagRerankTopK:           v.GetInt64("ragRerankTopK"),
		RagRerankScoreThreshold: v.GetFloat64("ragRerankScoreThreshold"),
		// ragSummaryRerankScoreThreshold 兼容：未配置时沿用 ragRerankScoreThreshold
		RagSummaryRerankScoreThreshold: func() float64 {
			if v.IsSet("ragSummaryRerankScoreThreshold") {
				return v.GetFloat64("ragSummaryRerankScoreThreshold")
			}
			return v.GetFloat64("ragRerankScoreThreshold")
		}(),
	}

	// 如果配置值为零值，设置默认值
	if config.MaxTokens == 0 {
		config.MaxTokens = 4096
	}
	if config.Model == "" {
		// 对话模型为空，读取第一个不带 embedding 的模型
		config.Model = utils.LLMModelStore.ResolveChatModel("")
	}
	if config.EmbeddingModel == "" {
		// Embedding 模型为空，读取第一个带 embedding 的模型
		config.EmbeddingModel = utils.LLMModelStore.ResolveEmbeddingModel("")
	}
	if config.Temperature == 0 {
		config.Temperature = 0.7
	}
	if config.TopK == 0 {
		config.TopK = 40
	}
	if config.TopP == 0 {
		config.TopP = 0.9
	}
	if config.RepeatPenalty == 0 {
		config.RepeatPenalty = 1.1
	}
	if config.RagEmbeddingTopK <= 0 {
		config.RagEmbeddingTopK = 10
	}
	if config.RagRerankTopK <= 0 {
		config.RagRerankTopK = 5
	}
	// 设置相似度阈值和重排分值阈值的默认值
	if config.RagSimilarityThreshold == 0 {
		config.RagSimilarityThreshold = 0.3
	}
	if config.RagSummarySimilarityThreshold == 0 {
		config.RagSummarySimilarityThreshold = 0.3
	}
	if config.RagRerankScoreThreshold == 0 {
		config.RagRerankScoreThreshold = 0.3
	}
	if config.RagSummaryRerankScoreThreshold == 0 {
		config.RagSummaryRerankScoreThreshold = 0.3
	}

	utils.SetEmbeddingModelOverride(clientId, config.EmbeddingModel)
	utils.SetTenantLLMOverride(clientId, utils.TenantLLMOverride{
		Model:            strings.TrimSpace(config.Model),
		EmbeddingModel:   strings.TrimSpace(config.EmbeddingModel),
		RerankerModel:    strings.TrimSpace(config.RerankerModel),
		CompletionURL:    strings.TrimSpace(config.CompletionUrl),
		CompletionType:   strings.TrimSpace(config.CompletionType),
		CompletionAPIKey: strings.TrimSpace(config.CompletionApiKey),
		EmbeddingsURL:    strings.TrimSpace(config.EmbeddingsUrl),
		EmbeddingsType:   strings.TrimSpace(config.EmbeddingsType),
		EmbeddingsAPIKey: strings.TrimSpace(config.EmbeddingsApiKey),
		RerankerURL:      strings.TrimSpace(config.RerankerUrl),
		RerankerType:     strings.TrimSpace(config.RerankerType),
		RerankerAPIKey:   strings.TrimSpace(config.RerankerApiKey),
	})
	return config, nil
}

func (l *LoadLLMSettingLogic) LoadLLMSetting() (resp *types.LLMSettingResponse, err error) {
	// 从 context 获取 clientId
	clientId, ok := l.ctx.Value("clientId").(string)
	if !ok {
		l.Errorf("Failed to get clientId from context")
		return &types.LLMSettingResponse{
			Response: types.Response{
				Code:    response.ServerErrorCode,
				Message: "Failed to get clientId",
			},
		}, nil
	}

	// 加载配置文件
	config, err := LoadAIConfig(clientId)
	if err != nil {
		l.Errorf("Failed to load AI config: %v", err)
		return &types.LLMSettingResponse{
			Response: types.Response{
				Code:    response.ServerErrorCode,
				Message: "Failed to load config: " + err.Error(),
			},
		}, nil
	}
	if strings.TrimSpace(config.CompletionUrl) == "" {
		config.CompletionUrl = strings.TrimSpace(l.svcCtx.Config.Llm.CompletionUrl)
	}
	if strings.TrimSpace(config.CompletionType) == "" {
		config.CompletionType = strings.TrimSpace(l.svcCtx.Config.Llm.CompletionType)
	}
	if strings.TrimSpace(config.EmbeddingsUrl) == "" {
		config.EmbeddingsUrl = strings.TrimSpace(l.svcCtx.Config.Rag.EmbeddingsUrl)
	}
	if strings.TrimSpace(config.EmbeddingsType) == "" {
		config.EmbeddingsType = strings.TrimSpace(l.svcCtx.Config.Rag.EmbeddingsType)
	}
	if strings.TrimSpace(config.RerankerUrl) == "" {
		config.RerankerUrl = strings.TrimSpace(l.svcCtx.Config.Rag.RerankerUrl)
	}
	if strings.TrimSpace(config.RerankerType) == "" {
		config.RerankerType = strings.TrimSpace(l.svcCtx.Config.Rag.RerankerType)
	}
	if strings.TrimSpace(config.RerankerModel) == "" {
		config.RerankerModel = utils.LLMModelStore.ResolveRerankModel("")
	}

	resp = &types.LLMSettingResponse{
		Response: types.Response{
			Code:    response.SuccessCode,
			Message: "success",
		},
		Data: config,
	}
	return resp, nil
}

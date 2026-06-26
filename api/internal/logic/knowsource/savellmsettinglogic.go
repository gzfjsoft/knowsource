package knowsource

import (
	"context"
	"strings"

	"knowsource/api/internal/svc"
	"knowsource/api/internal/types"
	"knowsource/api/internal/utils"
	"knowsource/common/response"

	"github.com/spf13/viper"
	"github.com/zeromicro/go-zero/core/logx"
)

type SaveLLMSettingLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 更新设置
func NewSaveLLMSettingLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SaveLLMSettingLogic {
	return &SaveLLMSettingLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// saveAIConfig 使用 viper 保存配置到 ai_{clientId}.yaml 文件
func saveAIConfig(clientId string, config *types.LLMSettingData) error {
	filePath, err := getAIConfigFilePath(clientId)
	if err != nil {
		return err
	}

	// 使用 viper 保存配置
	v := viper.New()
	v.SetConfigFile(filePath)
	v.SetConfigType("yaml")
	v.Set("maxSize", config.MaxTokens)
	v.Set("model", config.Model)
	v.Set("embeddingModel", config.EmbeddingModel)
	v.Set("completionUrl", config.CompletionUrl)
	v.Set("completionApiKey", config.CompletionApiKey)
	v.Set("completionType", config.CompletionType)
	v.Set("embeddingsUrl", config.EmbeddingsUrl)
	v.Set("embeddingsApiKey", config.EmbeddingsApiKey)
	v.Set("embeddingsType", config.EmbeddingsType)
	v.Set("rerankerUrl", config.RerankerUrl)
	v.Set("rerankerApiKey", config.RerankerApiKey)
	v.Set("rerankerType", config.RerankerType)
	v.Set("rerankerModel", config.RerankerModel)
	v.Set("temperature", config.Temperature)
	v.Set("topK", config.TopK)
	v.Set("topP", config.TopP)
	v.Set("repeatPenalty", config.RepeatPenalty)
	v.Set("ragEmbeddingTopK", config.RagEmbeddingTopK)
	v.Set("ragSimilarityThreshold", config.RagSimilarityThreshold)
	v.Set("ragSummarySimilarityThreshold", config.RagSummarySimilarityThreshold)
	v.Set("ragRerankTopK", config.RagRerankTopK)
	v.Set("ragRerankScoreThreshold", config.RagRerankScoreThreshold)
	v.Set("ragSummaryRerankScoreThreshold", config.RagSummaryRerankScoreThreshold)

	// 如果文件不存在，WriteConfigAs 会创建它
	if err := v.WriteConfigAs(filePath); err != nil {
		return err
	}

	logx.Infof("Saved AI config to: %s", filePath)
	return nil
}

func (l *SaveLLMSettingLogic) SaveLLMSetting(req *types.LLMSettingData) (resp *types.LLMSettingResponse, err error) {
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

	// 验证参数
	if req == nil {
		return &types.LLMSettingResponse{
			Response: types.Response{
				Code:    response.ParameterErrorCode,
				Message: "请求参数不能为空",
			},
		}, nil
	}

	utils.SetEmbeddingModelOverride(clientId, req.EmbeddingModel)
	utils.SetTenantLLMOverride(clientId, utils.TenantLLMOverride{
		Model:            strings.TrimSpace(req.Model),
		EmbeddingModel:   strings.TrimSpace(req.EmbeddingModel),
		RerankerModel:    strings.TrimSpace(req.RerankerModel),
		CompletionURL:    strings.TrimSpace(req.CompletionUrl),
		CompletionType:   strings.TrimSpace(req.CompletionType),
		CompletionAPIKey: strings.TrimSpace(req.CompletionApiKey),
		EmbeddingsURL:    strings.TrimSpace(req.EmbeddingsUrl),
		EmbeddingsType:   strings.TrimSpace(req.EmbeddingsType),
		EmbeddingsAPIKey: strings.TrimSpace(req.EmbeddingsApiKey),
		RerankerURL:      strings.TrimSpace(req.RerankerUrl),
		RerankerType:     strings.TrimSpace(req.RerankerType),
		RerankerAPIKey:   strings.TrimSpace(req.RerankerApiKey),
	})
	// 保存配置到文件
	if err := saveAIConfig(clientId, req); err != nil {
		l.Errorf("Failed to save AI config: %v", err)
		return &types.LLMSettingResponse{
			Response: types.Response{
				Code:    response.ServerErrorCode,
				Message: "Failed to save config: " + err.Error(),
			},
		}, nil
	}

	resp = &types.LLMSettingResponse{
		Response: types.Response{
			Code:    response.SuccessCode,
			Message: "success",
		},
		Data: req,
	}
	return resp, nil
}

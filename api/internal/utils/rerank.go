package utils

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/zeromicro/go-zero/core/logx"
)

// DefaultRerankModel vLLM 默认重排模型
const DefaultRerankModel = "Qwen3-Reranker-0.6B"

// RerankRequest 重排请求
type RerankRequest struct {
	Query     string   `json:"query"`
	Documents []string `json:"documents"`
	Model     string   `json:"model,omitempty"`
}

// RerankResult 单条重排结果（按 relevance_score 降序）
type RerankResult struct {
	Index          int     `json:"index"`
	Text           string  `json:"text"`
	RelevanceScore float64 `json:"relevance_score"`
}

// vLLM rerank API 响应结构
type rerankResponse struct {
	ID    string `json:"id"`
	Model string `json:"model"`
	Usage struct {
		TotalTokens int `json:"total_tokens"`
	} `json:"usage"`
	Results []struct {
		Index    int `json:"index"`
		Document struct {
			Text       string      `json:"text"`
			MultiModal interface{} `json:"multi_modal"`
		} `json:"document"`
		// 部分服务只返回 score，或同时返回两种字段
		RelevanceScore float64 `json:"relevance_score"`
		Score          float64 `json:"score"`
	} `json:"results"`
}

// RerankByType 根据 rerankType 调用对应的重排实现：llama.cpp 或 vllm（默认）
func RerankByType(ctx context.Context, baseURL string, rerankType string, req RerankRequest) ([]RerankResult, error) {
	return RerankByTypeWithAPIKey(ctx, baseURL, "", rerankType, req)
}

func RerankByTypeWithAPIKey(ctx context.Context, baseURL string, apiKey string, rerankType string, req RerankRequest) ([]RerankResult, error) {
	t := strings.ToLower(strings.TrimSpace(rerankType))
	if t == "llama.cpp" || t == "llamacpp" {
		return RerankLlamacppWithAPIKey(ctx, baseURL, apiKey, req)
	}
	return RerankVllmWithAPIKey(ctx, baseURL, apiKey, req)
}

// Rerank 调用 vLLM POST {baseURL}/v1/rerank，按相关性重排文档列表（向后兼容）
// 新代码建议使用 RerankByType 根据配置的 RerankerType 选择实现
func Rerank(ctx context.Context, baseURL string, req RerankRequest) ([]RerankResult, error) {
	return RerankVllm(ctx, baseURL, req)
}

func RerankVllm(ctx context.Context, baseURL string, req RerankRequest) ([]RerankResult, error) {
	return RerankVllmWithAPIKey(ctx, baseURL, "", req)
}

// RerankVllm 调用 vLLM POST {baseURL}/v1/rerank，按相关性重排文档列表
// baseURL 如 http://127.0.0.1:8022，model 为空时使用 DefaultRerankModel
func RerankVllmWithAPIKey(ctx context.Context, baseURL string, apiKey string, req RerankRequest) ([]RerankResult, error) {
	baseURL = strings.TrimSuffix(baseURL, "/")
	if baseURL == "" {
		return nil, fmt.Errorf("rerank baseURL 为空")
	}
	if req.Query == "" {
		return nil, fmt.Errorf("rerank query 为空")
	}
	if len(req.Documents) == 0 {
		return nil, fmt.Errorf("rerank documents 为空")
	}
	model := req.Model
	if model == "" {
		model = DefaultRerankModel
	}

	url := baseURL + "/v1/rerank"
	body := map[string]interface{}{
		"query":     req.Query,
		"documents": req.Documents,
		"model":     model,
	}
	jsonData, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("rerank 请求体序列化: %w", err)
	}
	logx.Infof("rerank request: url=%s model=%s docs=%d", url, model, len(req.Documents))

	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("rerank 创建请求: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	if strings.TrimSpace(apiKey) != "" {
		httpReq.Header.Set("Authorization", "Bearer "+strings.TrimSpace(apiKey))
	}

	client := &http.Client{Timeout: 60 * time.Second}
	resp, err := client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("rerank 请求: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("rerank 读取响应: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("rerank 请求失败 status=%d body=%s", resp.StatusCode, string(respBody))
	}

	var result rerankResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("rerank 解析响应: %w", err)
	}
	out := make([]RerankResult, 0, len(result.Results))
	for _, r := range result.Results {
		rel := r.RelevanceScore
		if rel == 0 && r.Score != 0 {
			rel = r.Score
		}
		out = append(out, RerankResult{
			Index:          r.Index,
			Text:           r.Document.Text,
			RelevanceScore: rel,
		})
	}
	return out, nil
}

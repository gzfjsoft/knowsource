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

// DefaultLlamacppRerankModel llama.cpp 默认重排模型
const DefaultLlamacppRerankModel = "Qwen3-reranker"

// llamacppRerankResponse llama.cpp rerank API 响应
// 返回 results: [{index, relevance_score}]，无 document.text，需用原始 documents 按 index 补全
type llamacppRerankResponse struct {
	Model   string `json:"model"`
	Object  string `json:"object"`
	Usage   interface{} `json:"usage"`
	Results []struct {
		Index          int     `json:"index"`
		RelevanceScore float64 `json:"relevance_score"`
	} `json:"results"`
}

// RerankLlamacpp 调用 llama.cpp POST {baseURL}/v1/rerank
// 请求格式: {"model":"Qwen3-reranker","query":"...","documents":["..."]}
// 响应格式: {"results":[{"index":0,"relevance_score":0.98},...]}，无 document 字段，用 documents[r.Index] 补全 Text
func RerankLlamacpp(ctx context.Context, baseURL string, req RerankRequest) ([]RerankResult, error) {
	return RerankLlamacppWithAPIKey(ctx, baseURL, "", req)
}

func RerankLlamacppWithAPIKey(ctx context.Context, baseURL string, apiKey string, req RerankRequest) ([]RerankResult, error) {
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
		model = DefaultLlamacppRerankModel
	}

	url := baseURL + "/v1/rerank"
	body := map[string]interface{}{
		"model":     model,
		"query":     req.Query,
		"documents": req.Documents,
	}
	jsonData, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("rerank 请求体序列化: %w", err)
	}
	logx.Infof("rerank llamacpp request: url=%s model=%s docs=%d", url, model, len(req.Documents))

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

	var result llamacppRerankResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("rerank 解析响应: %w", err)
	}
	out := make([]RerankResult, 0, len(result.Results))
	for _, r := range result.Results {
		text := ""
		if r.Index >= 0 && r.Index < len(req.Documents) {
			text = req.Documents[r.Index]
		}
		out = append(out, RerankResult{
			Index:          r.Index,
			Text:           text,
			RelevanceScore: r.RelevanceScore,
		})
	}
	return out, nil
}

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

// OneShotChatRequest OpenAI 兼容的单轮对话请求
type OneShotChatRequest struct {
	Model       string           `json:"model,omitempty"` // 模型名，来自配置 ai.yaml 的 model
	Messages    []OneShotMessage `json:"messages"`
	Stream      bool             `json:"stream"`
	Temperature float64          `json:"temperature,omitempty"`
	TopP        float64          `json:"top_p,omitempty"`
	Think       bool             `json:"think,omitempty"`
	MaxTokens   int64            `json:"max_tokens,omitempty"`
}

type OneShotMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// openAIChoice OpenAI 格式的 choice
type openAIChoice struct {
	Message struct {
		Content string `json:"content"`
	} `json:"message"`
}

type openAIResponse struct {
	Choices []openAIChoice `json:"choices"`
}

// ollamaMessageResponse Ollama 格式的 message
type ollamaMessageResponse struct {
	Message struct {
		Content string `json:"content"`
	} `json:"message"`
}

// ollamaChatRequest Ollama 原生 /api/chat 请求体
type ollamaChatRequest struct {
	Model    string           `json:"model"`
	Messages []OneShotMessage `json:"messages"`
	Think    bool             `json:"think"`
	Stream   bool             `json:"stream"`
}

// CallLLMOllamaOneShot 使用 Ollama 原生 POST /api/chat（stream: false）进行一次对话，返回 message.content。
func CallLLMOllamaOneShot(ctx context.Context, ollamaBaseURL string, model string, userPrompt string, think bool) (content string, err error) {
	return CallLLMOllamaOneShotWithAPIKey(ctx, ollamaBaseURL, "", model, userPrompt, think)
}

func CallLLMOllamaOneShotWithAPIKey(ctx context.Context, ollamaBaseURL string, apiKey string, model string, userPrompt string, think bool) (content string, err error) {
	model = strings.TrimSpace(model)
	if model == "" {
		return "", fmt.Errorf("model is empty")
	}
	ollamaBaseURL = strings.TrimSuffix(ollamaBaseURL, "/")
	url := ollamaBaseURL + "/api/chat"
	body := ollamaChatRequest{
		Model: model,
		Messages: []OneShotMessage{
			{Role: "user", Content: strings.TrimSpace(userPrompt)},
		},
		Think:  think,
		Stream: false,
	}
	bodyBytes, err := json.Marshal(body)
	if err != nil {
		return "", fmt.Errorf("marshal ollama request: %w", err)
	}
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(bodyBytes))
	if err != nil {
		return "", fmt.Errorf("new request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	if strings.TrimSpace(apiKey) != "" {
		req.Header.Set("Authorization", "Bearer "+strings.TrimSpace(apiKey))
	}
	client := &http.Client{Timeout: 5 * time.Minute}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("do request: %w", err)
	}
	defer resp.Body.Close()
	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("read body: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("ollama api status %d: %s", resp.StatusCode, string(raw))
	}
	var ollama ollamaMessageResponse
	if err := json.Unmarshal(raw, &ollama); err != nil {
		return "", fmt.Errorf("parse ollama response: %w", err)
	}
	if ollama.Message.Content == "" {
		logx.WithContext(ctx).Infof("ollama raw response (no content): %s", string(raw))
		return "", fmt.Errorf("ollama response has no content")
	}
	return strings.TrimSpace(ollama.Message.Content), nil
}

// CallLLMOneShot 调用 vLLM/Ollama 进行一次非流式对话，返回模型回复的文本内容。
// apiUrl 为 Llm.CompletionUrl（OpenAI 兼容 /v1/chat/completions）。
// model 为使用的模型名，通常来自 ai.yaml 配置；为空时返回错误。
func CallLLMOneShot(ctx context.Context, apiUrl string, model string, userPrompt string, temperature float64, maxTokens int64, think bool) (content string, err error) {
	return CallLLMOneShotWithAPIKey(ctx, apiUrl, "", model, userPrompt, temperature, maxTokens, think)
}

func CallLLMOneShotWithAPIKey(ctx context.Context, apiUrl string, apiKey string, model string, userPrompt string, temperature float64, maxTokens int64, think bool) (content string, err error) {
	if apiUrl == "" {
		return "", fmt.Errorf("apiUrl is empty")
	}
	model = strings.TrimSpace(model)
	if model == "" {
		return "", fmt.Errorf("model is empty")
	}
	reqBody := OneShotChatRequest{
		Model: model,
		Messages: []OneShotMessage{
			{Role: "user", Content: strings.TrimSpace(userPrompt)},
		},
		Stream:      false,
		Temperature: temperature,
		TopP:        0.9,
		Think:       think,
		MaxTokens:   maxTokens,
	}
	if maxTokens <= 0 {
		reqBody.MaxTokens = 512
	}
	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("marshal request: %w", err)
	}
	// apiUrl 应为完整地址（如 http://host/v1/chat/completions 或 Ollama 兼容端点）
	url := strings.TrimSuffix(apiUrl, "/")
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(bodyBytes))
	if err != nil {
		return "", fmt.Errorf("new request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	if strings.TrimSpace(apiKey) != "" {
		req.Header.Set("Authorization", "Bearer "+strings.TrimSpace(apiKey))
	}
	// LLM 调用可能比较慢，这里设置长一点的超时时间（例如 5 分钟）
	client := &http.Client{Timeout: 5 * time.Minute}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("do request: %w", err)
	}
	defer resp.Body.Close()
	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("read body: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("llm api status %d: %s", resp.StatusCode, string(raw))
	}
	// 先按 OpenAI 格式解析
	var openAI openAIResponse
	if err := json.Unmarshal(raw, &openAI); err == nil && len(openAI.Choices) > 0 && openAI.Choices[0].Message.Content != "" {
		return strings.TrimSpace(openAI.Choices[0].Message.Content), nil
	}
	// 再按 Ollama 格式解析（部分服务可能返回 Ollama 格式）
	var ollama ollamaMessageResponse
	if err := json.Unmarshal(raw, &ollama); err == nil && ollama.Message.Content != "" {
		return strings.TrimSpace(ollama.Message.Content), nil
	}
	logx.WithContext(ctx).Infof("llm raw response (no content parsed): %s", string(raw))
	return "", fmt.Errorf("could not parse llm response content")
}

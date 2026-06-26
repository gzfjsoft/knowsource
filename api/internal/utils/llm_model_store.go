package utils

import (
	"strings"
	"sync"
)

// LLMModelStore 内存中保存 vLLM 三个地址的模型 ID 列表，供 chat/embedding/rerank 解析使用。
// 程序启动时和每次 /sys/check 都会更新。
var LLMModelStore = &llmModelStore{}

type llmModelStore struct {
	mu               sync.RWMutex
	ChatModelIds     []string
	EmbeddingModelIds []string
	RerankModelIds   []string
}

// Update 更新三个服务的模型 ID 列表
func (s *llmModelStore) Update(chat, embedding, rerank []string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.ChatModelIds = copyStrings(chat)
	s.EmbeddingModelIds = copyStrings(embedding)
	s.RerankModelIds = copyStrings(rerank)
}

func copyStrings(src []string) []string {
	if src == nil {
		return nil
	}
	dst := make([]string, len(src))
	copy(dst, src)
	return dst
}

// ResolveChatModel 解析 chat 模型：若 model 在列表中则返回，否则选一个含 qwen3 且不含 embedding 的
func (s *llmModelStore) ResolveChatModel(model string) string {
	s.mu.RLock()
	ids := s.ChatModelIds
	s.mu.RUnlock()
	return resolveModel(ids, model, false, true) // 不含 embedding
}

// ResolveEmbeddingModel 解析 embedding 模型：若 model 在列表中则返回，否则选含 embedding 的 qwen3
func (s *llmModelStore) ResolveEmbeddingModel(model string) string {
	s.mu.RLock()
	ids := s.EmbeddingModelIds
	s.mu.RUnlock()
	return resolveEmbeddingModel(ids, model)
}

// ResolveRerankModel 解析 rerank 模型：若 model 在列表中则返回，否则选含 rerank 的 qwen3
func (s *llmModelStore) ResolveRerankModel(model string) string {
	s.mu.RLock()
	ids := s.RerankModelIds
	s.mu.RUnlock()
	return resolveRerankModel(ids, model)
}

// resolveModel 通用解析：ids 为可用模型列表，prefer 为优先匹配（含 qwen3），excludeEmbedding 为排除含 embedding 的
func resolveModel(ids []string, model string, preferQwen3, excludeEmbedding bool) string {
	model = strings.TrimSpace(model)
	if model != "" && contains(ids, model) {
		return model
	}
	// 从列表中选一个：含 qwen3，不含 embedding
	for _, id := range ids {
		lower := strings.ToLower(id)
		if excludeEmbedding && strings.Contains(lower, "embedding") {
			continue
		}
		if preferQwen3 && !strings.Contains(lower, "qwen3") {
			continue
		}
		return id
	}
	// 若 preferQwen3 为 false，再试不含 embedding 的任意模型
	if excludeEmbedding {
		for _, id := range ids {
			if !strings.Contains(strings.ToLower(id), "embedding") {
				return id
			}
		}
	}
	// 列表为空或都不匹配，返回默认
	if model != "" {
		return model
	}
	return "Qwen3-0.6B"
}

func resolveEmbeddingModel(ids []string, model string) string {
	model = strings.TrimSpace(model)
	if model != "" && contains(ids, model) {
		return model
	}
	for _, id := range ids {
		lower := strings.ToLower(id)
		if strings.Contains(lower, "embedding") && strings.Contains(lower, "qwen") {
			return id
		}
	}
	for _, id := range ids {
		if strings.Contains(strings.ToLower(id), "embedding") {
			return id
		}
	}
	if model != "" {
		return model
	}
	return "Qwen3-Embedding-0.6B"
}

func resolveRerankModel(ids []string, model string) string {
	model = strings.TrimSpace(model)
	if model != "" && contains(ids, model) {
		return model
	}
	for _, id := range ids {
		lower := strings.ToLower(id)
		if strings.Contains(lower, "rerank") && strings.Contains(lower, "qwen") {
			return id
		}
	}
	for _, id := range ids {
		if strings.Contains(strings.ToLower(id), "rerank") {
			return id
		}
	}
	if model != "" {
		return model
	}
	return "Qwen3-Reranker-0.6B"
}

func contains(slice []string, s string) bool {
	for _, v := range slice {
		if v == s {
			return true
		}
	}
	return false
}

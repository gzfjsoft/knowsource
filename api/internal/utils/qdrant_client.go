package utils

import (
	"bytes"
	"context"
	"crypto/md5"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/qdrant/go-client/qdrant"
	"github.com/zeromicro/go-zero/core/logx"
)

// QdrantTools Qdrant 向量数据库客户端
type QdrantTools struct {
	HTTPClient          *http.Client
	client              *qdrant.Client
	EmbeddingAPI        string // Ollama Embedding API 地址（可选）
	EmbeddingAPIKey     string // OpenAI 兼容鉴权 Key（可选）
	EmbeddingModel      string // Embedding 模型名（可选，不设置则走默认）
	VllmEmbeddingURL    string // vLLM Embedding 地址，如 http://localhost:8021，优先于 EmbeddingAPI
	VllmEmbeddingAPIKey string // vLLM/OpenAI 兼容鉴权 Key（可选）
	VllmEmbeddingModel  string // vLLM 向量模型路径，如 Qwen3-Embedding-0.6B
}

// FindDocFilenamesByRerank 对“答案”与候选文件名做重排打分，返回分值 >= threshold 的文件名（按分值降序）。
// 说明：
// - 这是一个“归因/定位参考来源”的启发式方法：用 rerank 判断答案更像引用了哪些文件名。
// - 依赖配置的 reranker（baseURL/type）；未配置时会返回空列表且不报错。
func (qc *QdrantTools) FindDocFilenamesByRerank(ctx context.Context, answer string, filenames []string, baseURL string, rerankType string, threshold float64) ([]string, []RerankResult, error) {
	return qc.FindDocFilenamesByRerankWithAPIKey(ctx, answer, filenames, baseURL, "", rerankType, threshold)
}

func (qc *QdrantTools) FindDocFilenamesByRerankWithAPIKey(ctx context.Context, answer string, filenames []string, baseURL string, apiKey string, rerankType string, threshold float64) ([]string, []RerankResult, error) {
	answer = strings.TrimSpace(answer)
	if answer == "" {
		return nil, nil, nil
	}
	baseURL = strings.TrimSpace(baseURL)
	if baseURL == "" {
		return nil, nil, nil
	}
	if threshold <= 0 {
		threshold = 0.8
	}

	candidates := make([]string, 0, len(filenames))
	seen := make(map[string]struct{}, len(filenames))
	for _, fn := range filenames {
		fn = strings.TrimSpace(fn)
		if fn == "" {
			continue
		}
		if _, ok := seen[fn]; ok {
			continue
		}
		seen[fn] = struct{}{}
		candidates = append(candidates, fn)
	}
	if len(candidates) == 0 {
		return nil, nil, nil
	}

	results, err := RerankByTypeWithAPIKey(ctx, baseURL, apiKey, rerankType, RerankRequest{
		Query:     answer,
		Documents: candidates,
		Model:     LLMModelStore.ResolveRerankModel(""),
	})
	if err != nil {
		return nil, nil, err
	}
	// 保险起见按分值降序排序（不同 rerank 服务未必保证有序返回）
	sort.Slice(results, func(i, j int) bool {
		return results[i].RelevanceScore > results[j].RelevanceScore
	})

	matched := make([]string, 0, len(results))
	for _, r := range results {
		if r.Index < 0 || r.Index >= len(candidates) {
			continue
		}
		if r.RelevanceScore >= threshold {
			matched = append(matched, candidates[r.Index])
		}
	}
	// 若无任何候选达到阈值，则回退返回 top1（最高分）文件名，保证至少有一个归因结果
	if len(matched) == 0 && len(results) > 0 {
		top := results[0]
		if top.Index >= 0 && top.Index < len(candidates) {
			matched = append(matched, candidates[top.Index])
		}
	}
	return matched, results, nil
}

// Point Qdrant 点（文档块）
type Point struct {
	ID      string                 `json:"id"`
	Vector  []float32              `json:"vector"`
	Payload map[string]interface{} `json:"payload"`
}

// CollectionInfo 集合信息
type CollectionInfo struct {
	Name       string `json:"name"`
	VectorSize int    `json:"vector_size"`
	Distance   string `json:"distance"` // "Cosine", "Euclid", "Dot"
}

// KeywordScrollResult 关键字滚动检索返回结构
type KeywordScrollResult struct {
	ID      interface{}            `json:"id"`
	Score   float32                `json:"score"`
	Payload map[string]interface{} `json:"payload"`
}

// QdrantScrollPoint Scroll 得到的点（id + payload），供业务层解析 metadata / page_content
type QdrantScrollPoint struct {
	ID      interface{}
	Payload map[string]interface{}
}

// NewQdrantTools 创建新的 Qdrant 客户端
func NewQdrantTools(host string, port int) *QdrantTools {
	// baseURL := fmt.Sprintf("http://%s:%d", host, port)

	httpClient := &http.Client{
		Timeout: 60 * time.Second,
	}

	// 使用官方 qdrant go-client 创建 gRPC 客户端
	qClient, err := qdrant.NewClient(&qdrant.Config{
		Host: host,
		Port: port,
		// 目前项目里本地部署，未启用 TLS / API Key，如后续需要可从配置中注入
	})
	if err != nil {
		logx.Errorf("failed to create qdrant client via go-client: %v", err)
	}

	return &QdrantTools{

		HTTPClient: httpClient,
		client:     qClient,
	}
}

// NewQdrantToolsWithClient 使用已有 qdrant client 构建工具实例
func NewQdrantToolsWithClient(client *qdrant.Client) *QdrantTools {
	httpClient := &http.Client{
		Timeout: 60 * time.Second,
	}
	return &QdrantTools{
		HTTPClient: httpClient,
		client:     client,
	}
}

// SetEmbeddingAPI 设置 Ollama Embedding API 地址
func (qc *QdrantTools) SetEmbeddingAPI(apiURL string) {
	qc.EmbeddingAPI = apiURL
}

func (qc *QdrantTools) SetEmbeddingAPIKey(apiKey string) {
	qc.EmbeddingAPIKey = strings.TrimSpace(apiKey)
}

// SetEmbeddingModel 设置 Embedding 模型名（可选）
func (qc *QdrantTools) SetEmbeddingModel(model string) {
	qc.EmbeddingModel = strings.TrimSpace(model)
}

// DefaultVllmEmbeddingModel vLLM 默认向量模型路径

const DefaultVllmEmbeddingModel = "Qwen3-Embedding-0.6B"

// SetVllmEmbedding 设置 vLLM Embedding 地址与模型（POST /v1/embeddings）
func (qc *QdrantTools) SetVllmEmbedding(baseURL, model string) {
	qc.VllmEmbeddingURL = strings.TrimSuffix(baseURL, "/")
	if model != "" {
		qc.VllmEmbeddingModel = model
	} else {
		qc.VllmEmbeddingModel = DefaultVllmEmbeddingModel
	}
}

func (qc *QdrantTools) SetVllmEmbeddingAPIKey(apiKey string) {
	qc.VllmEmbeddingAPIKey = strings.TrimSpace(apiKey)
}

// CollectionExists 检查集合是否存在
func (qc *QdrantTools) CollectionExists(ctx context.Context, collectionName string) (bool, error) {
	if qc.client == nil {
		return false, fmt.Errorf("Qdrant 客户端未初始化")
	}
	return qc.client.CollectionExists(ctx, collectionName)
}

// ListCollections 列出当前 Qdrant 中的 collection 名称
func (qc *QdrantTools) ListCollections(ctx context.Context) ([]string, error) {
	if qc.client == nil {
		return nil, fmt.Errorf("Qdrant 客户端未初始化")
	}
	resp, err := qc.client.ListCollections(ctx)
	if err != nil {
		return nil, fmt.Errorf("获取 collection 列表失败: %w", err)
	}
	names := make([]string, 0, len(resp))
	for _, item := range resp {
		if item != "" {
			names = append(names, item)
		}
	}
	return names, nil
}

// CreateCollection 创建集合
func (qc *QdrantTools) CreateCollection(ctx context.Context, collectionName string, vectorSize int, distance string) error {
	if qc.client == nil {
		return fmt.Errorf("Qdrant 客户端未初始化")
	}

	var dist qdrant.Distance
	switch strings.ToLower(distance) {
	case "cosine":
		dist = qdrant.Distance_Cosine
	case "euclid":
		dist = qdrant.Distance_Euclid
	case "dot":
		dist = qdrant.Distance_Dot
	default:
		dist = qdrant.Distance_Dot
	}

	req := &qdrant.CreateCollection{
		CollectionName: collectionName,
		VectorsConfig: qdrant.NewVectorsConfig(&qdrant.VectorParams{
			Size:     uint64(vectorSize),
			Distance: dist,
		}),
	}

	if err := qc.client.CreateCollection(ctx, req); err != nil {
		return fmt.Errorf("创建集合失败（go-client）: %w", err)
	}
	return nil
}

// CreatePayloadIndex 创建 payload 索引
func (qc *QdrantTools) CreatePayloadIndex(ctx context.Context, collectionName string, fieldName string, fieldSchema string) error {
	if qc.client == nil {
		return fmt.Errorf("Qdrant 客户端未初始化")
	}

	var ft qdrant.FieldType
	switch strings.ToLower(fieldSchema) {
	case "keyword":
		ft = qdrant.FieldType_FieldTypeKeyword
	case "integer", "int":
		ft = qdrant.FieldType_FieldTypeInteger
	case "float", "double":
		ft = qdrant.FieldType_FieldTypeFloat
	case "text":
		ft = qdrant.FieldType_FieldTypeText
	case "bool", "boolean":
		ft = qdrant.FieldType_FieldTypeBool
	case "datetime":
		ft = qdrant.FieldType_FieldTypeDatetime
	case "uuid":
		ft = qdrant.FieldType_FieldTypeUuid
	default:
		// 默认 keyword
		ft = qdrant.FieldType_FieldTypeKeyword
	}

	wait := true
	req := &qdrant.CreateFieldIndexCollection{
		CollectionName: collectionName,
		Wait:           &wait,
		FieldName:      fieldName,
		FieldType:      &ft,
	}

	if _, err := qc.client.CreateFieldIndex(ctx, req); err != nil {
		return fmt.Errorf("创建索引失败（go-client）: %w", err)
	}
	return nil
}

// DefaultEmbeddingModel 默认向量模型
const DefaultEmbeddingModel = "qwen3-embedding:4b"

// DefaultEmbeddingDimensions 默认向量维度（可选，0 表示不传）

// GenerateEmbedding 生成文本的嵌入向量。优先使用 vLLM POST /v1/embeddings，否则使用 Ollama POST /api/embed
func (qc *QdrantTools) GenerateEmbedding(ctx context.Context, text string) ([]float32, error) {
	if qc.VllmEmbeddingURL != "" {
		return qc.callVllmEmbedAPI(ctx, text)
	}
	if qc.EmbeddingAPI != "" {
		return qc.callEmbedAPI(ctx, text)
	}
	return nil, fmt.Errorf("未配置 Embedding API（Rag.EmbeddingsUrl 或 Ollama），无法生成向量")
}

// embedRequest /api/embed 请求体
type embedRequest struct {
	Model string   `json:"model"`
	Input []string `json:"input"`
	// Options *embedOption `json:"options,omitempty"`
}

// type embedOption struct {
// 	Dimensions int `json:"dimensions,omitempty"`
// }

// embedResponseData 兼容两种常见返回：data[].embedding 或 直接 embedding
type embedResponseData struct {
	Embedding []float64 `json:"embedding"`
}

// embedResponse /api/embed 响应：embeddings 或 data 或 embedding
// 实际返回: {"model":"...","embeddings":[[...]],"total_duration":...}
type embedResponse struct {
	Embeddings [][]float64         `json:"embeddings"` //   等: 多条向量 [][]
	Data       []embedResponseData `json:"data"`       // OpenAI 风格
	Embedding  []float64           `json:"embedding"`  // 单条直接返回
}

// callEmbedAPI 调用 POST {EmbeddingAPI}/api/embed
// body: {"model":"qwen3-embedding:0.6b","input":["文本"],"options":{"dimensions":1024}}
func (qc *QdrantTools) callEmbedAPI(ctx context.Context, text string) ([]float32, error) {
	url := qc.EmbeddingAPI
	if len(url) > 0 && url[len(url)-1] == '/' {
		url = url[:len(url)-1]
	}
	url = url + "/api/embed"

	// 优先使用显式注入的 embedding model；未注入时使用默认值
	model := strings.TrimSpace(qc.EmbeddingModel)
	if model == "" {
		model = DefaultEmbeddingModel
	}

	body := embedRequest{
		Model: model,
		Input: []string{text},
	}
	// if DefaultEmbeddingDimensions > 0 {
	// 	body.Options = &embedOption{Dimensions: DefaultEmbeddingDimensions}
	// }
	jsonData, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}

	logx.Infof("embed request: %s, url: %s", string(jsonData), url)

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	if strings.TrimSpace(qc.EmbeddingAPIKey) != "" {
		req.Header.Set("Authorization", "Bearer "+strings.TrimSpace(qc.EmbeddingAPIKey))
	}

	resp, err := qc.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("embed 请求失败 status=%d body=%s", resp.StatusCode, string(respBody))
	}

	var result embedResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("解析 embed 响应: %w", err)
	}

	logx.Infof("embed response: %s", string(respBody))

	var raw []float64
	if len(result.Embeddings) > 0 && len(result.Embeddings[0]) > 0 {
		raw = result.Embeddings[0]
	} else if len(result.Data) > 0 && len(result.Data[0].Embedding) > 0 {
		raw = result.Data[0].Embedding
	} else if len(result.Embedding) > 0 {
		raw = result.Embedding
	}
	if len(raw) == 0 {
		return nil, fmt.Errorf("embed 返回空向量")
	}

	vec := make([]float32, len(raw))
	for i, v := range raw {
		vec[i] = float32(v)
	}
	return vec, nil
}

// vllmEmbedRequest /v1/embeddings 请求体（vLLM/OpenAI 兼容）
type vllmEmbedRequest struct {
	Input string `json:"input"`
	Model string `json:"model,omitempty"`
}

// callVllmEmbedAPI 调用 vLLM POST {base}/v1/embeddings，请求体: {"input":"文本","model":"模型路径"}
func (qc *QdrantTools) callVllmEmbedAPI(ctx context.Context, text string) ([]float32, error) {
	url := qc.VllmEmbeddingURL + "/v1/embeddings"
	model := qc.VllmEmbeddingModel
	if model == "" {
		model = DefaultVllmEmbeddingModel
	}
	body := vllmEmbedRequest{Input: text, Model: model}
	jsonData, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	logx.Infof("vllm embed request: url=%s  ", url)

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	if strings.TrimSpace(qc.VllmEmbeddingAPIKey) != "" {
		req.Header.Set("Authorization", "Bearer "+strings.TrimSpace(qc.VllmEmbeddingAPIKey))
	}

	resp, err := qc.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("vllm embed 请求失败 status=%d body=%s", resp.StatusCode, string(respBody))
	}

	var result embedResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("解析 vllm embed 响应: %w", err)
	}
	var raw []float64
	if len(result.Data) > 0 && len(result.Data[0].Embedding) > 0 {
		raw = result.Data[0].Embedding
	} else if len(result.Embeddings) > 0 && len(result.Embeddings[0]) > 0 {
		raw = result.Embeddings[0]
	} else if len(result.Embedding) > 0 {
		raw = result.Embedding
	}
	if len(raw) == 0 {
		return nil, fmt.Errorf("vllm embed 返回空向量")
	}
	logx.Infof("vllm embed 长度: %v", len(raw))
	vec := make([]float32, len(raw))
	for i, v := range raw {
		vec[i] = float32(v)
	}
	return vec, nil
}

// UpsertPoints 批量插入或更新点
func (qc *QdrantTools) UpsertPoints(ctx context.Context, collectionName string, points []Point) error {
	if qc.client == nil {
		return fmt.Errorf("Qdrant 客户端未初始化")
	}

	if len(points) == 0 {
		return nil
	}

	upsertPoints := make([]*qdrant.PointStruct, 0, len(points))
	for _, p := range points {
		// 将自定义 Point 转为 qdrant.PointStruct
		vec := make([]float32, len(p.Vector))
		copy(vec, p.Vector)

		ps := &qdrant.PointStruct{
			Id:      qdrant.NewID(p.ID),
			Vectors: qdrant.NewVectors(vec...),
			Payload: qdrant.NewValueMap(p.Payload),
		}
		upsertPoints = append(upsertPoints, ps)
	}

	wait := true
	_, err := qc.client.Upsert(ctx, &qdrant.UpsertPoints{
		CollectionName: collectionName,
		Wait:           &wait,
		Points:         upsertPoints,
	})
	if err != nil {
		return fmt.Errorf("Qdrant Upsert 失败: %w", err)
	}
	return nil
}

// ScoredPoint 检索返回的单条结果（Qdrant points/search）
type ScoredPoint struct {
	ID      interface{}            `json:"id"`
	Score   float64                `json:"score"`
	Payload map[string]interface{} `json:"payload"`
}

// searchPointsRequest Qdrant POST /collections/{name}/points/search 请求体
type searchPointsRequest struct {
	Vector      []float32              `json:"vector"`
	Limit       int                    `json:"limit"`
	WithPayload bool                   `json:"with_payload"`
	WithVector  bool                   `json:"with_vector"`
	Filter      map[string]interface{} `json:"filter,omitempty"`
}

// searchPointsResponse 响应 result 数组元素
type searchPointsResultItem struct {
	ID      interface{}            `json:"id"`
	Score   float64                `json:"score"`
	Payload map[string]interface{} `json:"payload"`
}

type searchPointsResponse struct {
	Result []searchPointsResultItem `json:"result"`
}

const qdrantSDKScrollPageLimit = 256

// extractQdrantValueToInterface 从 qdrant.Value 提取为 interface{}
func extractQdrantValueToInterface(value *qdrant.Value) interface{} {
	if value == nil {
		return nil
	}
	switch v := value.Kind.(type) {
	case *qdrant.Value_StringValue:
		return v.StringValue
	case *qdrant.Value_IntegerValue:
		return v.IntegerValue
	case *qdrant.Value_DoubleValue:
		return v.DoubleValue
	case *qdrant.Value_BoolValue:
		return v.BoolValue
	case *qdrant.Value_ListValue:
		var list []interface{}
		for _, item := range v.ListValue.Values {
			list = append(list, extractQdrantValueToInterface(item))
		}
		return list
	case *qdrant.Value_StructValue:
		m := make(map[string]interface{})
		for key, val := range v.StructValue.Fields {
			m[key] = extractQdrantValueToInterface(val)
		}
		return m
	default:
		return nil
	}
}

// extractQdrantIDToInterface 从 PointId 提取为 interface{}
func extractQdrantIDToInterface(id *qdrant.PointId) interface{} {
	if id == nil {
		return nil
	}
	switch v := id.PointIdOptions.(type) {
	case *qdrant.PointId_Num:
		return v.Num
	case *qdrant.PointId_Uuid:
		return v.Uuid
	default:
		return nil
	}
}

// scrollExtractQdrantValue 从 qdrant.Value 转为 Go 值（供 scroll 结果解析）
func scrollExtractQdrantValue(value *qdrant.Value) interface{} {
	if value == nil {
		return nil
	}
	switch v := value.Kind.(type) {
	case *qdrant.Value_StringValue:
		return v.StringValue
	case *qdrant.Value_IntegerValue:
		return v.IntegerValue
	case *qdrant.Value_DoubleValue:
		return v.DoubleValue
	case *qdrant.Value_BoolValue:
		return v.BoolValue
	case *qdrant.Value_ListValue:
		var list []interface{}
		for _, item := range v.ListValue.Values {
			list = append(list, scrollExtractQdrantValue(item))
		}
		return list
	case *qdrant.Value_StructValue:
		m := make(map[string]interface{})
		for key, val := range v.StructValue.Fields {
			m[key] = scrollExtractQdrantValue(val)
		}
		return m
	default:
		return value
	}
}

// scrollExtractQdrantID 从 PointId 提取可 JSON 序列化的 id
func scrollExtractQdrantID(id *qdrant.PointId) interface{} {
	if id == nil {
		return nil
	}
	switch v := id.PointIdOptions.(type) {
	case *qdrant.PointId_Num:
		return v.Num
	case *qdrant.PointId_Uuid:
		return v.Uuid
	default:
		return id
	}
}

// searchPointsSDK 使用 qdrant SDK 进行向量检索，返回按相似度排序的 top limit 条。
// tags 非空时只检索 metadata.tag 匹配任意一个 tag 的点（OR）
func searchPointsSDK(ctx context.Context, client *qdrant.Client, collectionName string, vector []float32, limit int, tags []string) ([]ScoredPoint, error) {
	return searchPointsWithFilterSDK(ctx, client, collectionName, vector, limit, tags, nil)
}

// searchPointsWithFilterSDK 使用 qdrant SDK 进行向量检索，支持标签和文件名过滤
func searchPointsWithFilterSDK(ctx context.Context, client *qdrant.Client, collectionName string, vector []float32, limit int, tags []string, fileNames []string) ([]ScoredPoint, error) {
	if client == nil {
		return nil, fmt.Errorf("Qdrant 客户端未初始化")
	}
	if limit <= 0 {
		limit = 10
	}

	var conditions []*qdrant.Condition

	// 标签过滤：metadata.tag 匹配任意一个
	if len(tags) > 0 {
		var tagList []string
		for _, t := range tags {
			if s := strings.TrimSpace(t); s != "" {
				tagList = append(tagList, s)
			}
		}
		if len(tagList) > 0 {
			conditions = append(conditions, qdrant.NewMatchKeywords("metadata.tag", tagList...))
		}
	}

	// 文件名过滤：metadata.file_name 匹配任意一个
	if len(fileNames) > 0 {
		var names []string
		for _, f := range fileNames {
			if s := strings.TrimSpace(f); s != "" {
				names = append(names, s)
			}
		}
		if len(names) > 0 {
			conditions = append(conditions, qdrant.NewMatchKeywords("metadata.file_name", names...))
		}
	}

	var filter *qdrant.Filter
	if len(conditions) > 0 {
		filter = &qdrant.Filter{Must: conditions}
	}

	limitU := uint64(limit)
	req := &qdrant.QueryPoints{
		CollectionName: collectionName,
		Query:          qdrant.NewQueryNearest(qdrant.NewVectorInputDense(vector)),
		Filter:         filter,
		Limit:          &limitU,
		WithPayload:    qdrant.NewWithPayload(true),
		WithVectors:    qdrant.NewWithVectors(false),
	}

	results, err := client.Query(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("Qdrant 检索失败: %w", err)
	}

	out := make([]ScoredPoint, 0, len(results))
	for _, r := range results {
		payload := make(map[string]interface{})
		if r.Payload != nil {
			for k, v := range r.Payload {
				payload[k] = extractQdrantValueToInterface(v)
			}
		}
		var id interface{}
		if r.Id != nil {
			id = extractQdrantIDToInterface(r.Id)
		}
		out = append(out, ScoredPoint{
			ID:      id,
			Score:   float64(r.Score),
			Payload: payload,
		})
	}
	return out, nil
}

// scrollPointsByFileNameSDK 使用 Qdrant 官方 gRPC SDK 的 Scroll，按 metadata.file_name 精确匹配拉取全部点。
// client 为 nil 返回错误；collection 不存在时返回 (nil, nil)。
func scrollPointsByFileNameSDK(ctx context.Context, client *qdrant.Client, collectionName, fileName string) ([]QdrantScrollPoint, error) {
	if client == nil {
		return nil, fmt.Errorf("Qdrant 客户端未初始化")
	}
	if collectionName == "" || fileName == "" {
		return nil, fmt.Errorf("collection 或 fileName 为空")
	}

	scrollFilter := &qdrant.Filter{
		Must: []*qdrant.Condition{qdrant.NewMatch("metadata.file_name", fileName)},
	}

	limit := uint32(qdrantSDKScrollPageLimit)
	var nextPageOffset *qdrant.PointId
	var all []QdrantScrollPoint

	for {
		scrollResult, err := client.Scroll(ctx, &qdrant.ScrollPoints{
			CollectionName: collectionName,
			Filter:         scrollFilter,
			Offset:         nextPageOffset,
			Limit:          &limit,
			WithPayload:    qdrant.NewWithPayload(true),
			WithVectors:    qdrant.NewWithVectors(false),
		})
		if err != nil {
			errStr := strings.ToLower(err.Error())
			if strings.Contains(errStr, "doesn't exist") ||
				strings.Contains(errStr, "not found") ||
				strings.Contains(errStr, "404") {
				return nil, nil
			}
			return nil, fmt.Errorf("Qdrant Scroll 失败: %w", err)
		}
		if len(scrollResult) == 0 {
			break
		}

		for _, point := range scrollResult {
			payload := make(map[string]interface{})
			if point.Payload != nil {
				for key, value := range point.Payload {
					payload[key] = scrollExtractQdrantValue(value)
				}
			}
			var id interface{}
			if point.Id != nil {
				id = scrollExtractQdrantID(point.Id)
			}
			all = append(all, QdrantScrollPoint{
				ID:      id,
				Payload: payload,
			})
		}

		last := scrollResult[len(scrollResult)-1]
		if last.Id != nil {
			nextPageOffset = last.Id
		}

		if len(scrollResult) < int(limit) {
			break
		}
	}

	return all, nil
}

// buildQdrantFilterFromMap 将简单的 map 结构转换为 qdrant.Filter
// 目前仅支持形如：
// {"must": []map[string]interface{}{ {"key":"metadata.file_name","match":{"value":"xxx"}} }}
func buildQdrantFilterFromMap(filter map[string]interface{}) (*qdrant.Filter, error) {
	if filter == nil {
		return nil, nil
	}

	rawMust, ok := filter["must"]
	if !ok {
		return nil, nil
	}

	mustSlice, ok := rawMust.([]map[string]interface{})
	if !ok {
		// 兼容 interface{} 切片
		if ifaceSlice, ok2 := rawMust.([]interface{}); ok2 {
			conds := make([]*qdrant.Condition, 0, len(ifaceSlice))
			for _, item := range ifaceSlice {
				m, ok3 := item.(map[string]interface{})
				if !ok3 {
					continue
				}
				cond, err := buildConditionFromMap(m)
				if err != nil {
					return nil, err
				}
				if cond != nil {
					conds = append(conds, cond)
				}
			}
			if len(conds) == 0 {
				return nil, nil
			}
			return &qdrant.Filter{Must: conds}, nil
		}
		return nil, fmt.Errorf("不支持的 filter.must 类型")
	}

	conds := make([]*qdrant.Condition, 0, len(mustSlice))
	for _, m := range mustSlice {
		cond, err := buildConditionFromMap(m)
		if err != nil {
			return nil, err
		}
		if cond != nil {
			conds = append(conds, cond)
		}
	}
	if len(conds) == 0 {
		return nil, nil
	}
	return &qdrant.Filter{Must: conds}, nil
}

// buildConditionFromMap 支持简单的 key + match.value / match.any 形式
func buildConditionFromMap(m map[string]interface{}) (*qdrant.Condition, error) {
	key, _ := m["key"].(string)
	if key == "" {
		return nil, nil
	}

	matchRaw, ok := m["match"].(map[string]interface{})
	if !ok {
		return nil, nil
	}

	// 单值匹配：{"value": "xxx"}
	if v, ok := matchRaw["value"]; ok {
		if val, ok := v.(string); ok {
			return qdrant.NewMatchText(key, val), nil
		}
		return nil, nil
	}

	// 多值匹配：{"any": [...]}
	if anyRaw, ok := matchRaw["any"]; ok {
		switch list := anyRaw.(type) {
		case []string:
			return qdrant.NewMatchKeywords(key, list...), nil
		case []interface{}:
			var strs []string
			for _, item := range list {
				if s, ok := item.(string); ok {
					strs = append(strs, s)
				}
			}
			if len(strs) > 0 {
				return qdrant.NewMatchKeywords(key, strs...), nil
			}
		}
	}

	return nil, nil
}

// SearchPoints 向量检索，返回按相似度排序的 top limit 条。tags 非空时只检索 metadata.tag 匹配任意一个 tag 的点（OR）
func (qc *QdrantTools) SearchPoints(ctx context.Context, collectionName string, vector []float32, limit int, tags []string) ([]ScoredPoint, error) {
	if qc.client == nil {
		return nil, fmt.Errorf("Qdrant 客户端未初始化")
	}
	return searchPointsSDK(ctx, qc.client, collectionName, vector, limit, tags)
}

// SearchPointsWithFileFilter 向量检索，支持 tags + fileNames 过滤
func (qc *QdrantTools) SearchPointsWithFileFilter(ctx context.Context, collectionName string, vector []float32, limit int, tags []string, fileNames []string) ([]ScoredPoint, error) {
	if qc.client == nil {
		return nil, fmt.Errorf("Qdrant 客户端未初始化")
	}
	return searchPointsWithFilterSDK(ctx, qc.client, collectionName, vector, limit, tags, fileNames)
}

// _SearchPointsWithFilter 预留：如果后续需要自定义 Filter，可以基于 SDK 实现（当前未使用）
func (qc *QdrantTools) _SearchPointsWithFilter(ctx context.Context, collectionName string, vector []float32, limit int, tags []string, customFilter map[string]interface{}) ([]ScoredPoint, error) {
	_ = customFilter
	return qc.SearchPoints(ctx, collectionName, vector, limit, tags)
}

// DeletePointsByFilter 按条件删除点。filter 为 Qdrant Filter 结构，如 map[string]interface{}{"must": []map[string]interface{}{{"key": "metadata.file_name", "match": map[string]interface{}{"value": "xxx"}}}}
func (qc *QdrantTools) DeletePointsByFilter(ctx context.Context, collectionName string, filter map[string]interface{}) error {
	if qc.client == nil {
		return fmt.Errorf("Qdrant 客户端未初始化")
	}

	qFilter, err := buildQdrantFilterFromMap(filter)
	if err != nil {
		return err
	}

	wait := true
	_, err = qc.client.Delete(ctx, &qdrant.DeletePoints{
		CollectionName: collectionName,
		Wait:           &wait,
		Points: &qdrant.PointsSelector{
			PointsSelectorOneOf: &qdrant.PointsSelector_Filter{
				Filter: qFilter,
			},
		},
	})
	if err != nil {
		return fmt.Errorf("按条件删除点失败: %w", err)
	}
	return nil
}

// DeletePointsByFileName 删除指定集合中 metadata.file_name 等于 fileName 的所有点
func (qc *QdrantTools) DeletePointsByFileName(ctx context.Context, collectionName string, fileName string) error {
	filter := map[string]interface{}{
		"must": []map[string]interface{}{
			{
				"key":   "metadata.file_name",
				"match": map[string]interface{}{"value": fileName},
			},
		},
	}
	return qc.DeletePointsByFilter(ctx, collectionName, filter)
}

// ScrollByKeyword 按 tags 先过滤，再在 payload 文本中做关键字匹配（大小写不敏感）
func (qc *QdrantTools) ScrollByKeyword(ctx context.Context, collectionName string, keyword string, tags []string, maxResults int) ([]KeywordScrollResult, error) {
	if qc.client == nil {
		return nil, fmt.Errorf("Qdrant 客户端未初始化")
	}
	if strings.TrimSpace(collectionName) == "" {
		return nil, fmt.Errorf("collectionName 不能为空")
	}
	if maxResults <= 0 {
		maxResults = 20
	}

	var results []KeywordScrollResult
	var scrollFilter *qdrant.Filter
	var tagList []string
	for _, t := range tags {
		if s := strings.TrimSpace(t); s != "" {
			tagList = append(tagList, s)
		}
	}
	if len(tagList) > 0 {
		scrollFilter = &qdrant.Filter{
			Must: []*qdrant.Condition{qdrant.NewMatchKeywords("metadata.tag", tagList...)},
		}
	}

	limit := uint32(100)
	var nextPageOffset *qdrant.PointId
	lowerKeyword := strings.ToLower(strings.TrimSpace(keyword))

	for len(results) < maxResults {
		scrollResult, err := qc.client.Scroll(ctx, &qdrant.ScrollPoints{
			CollectionName: collectionName,
			Filter:         scrollFilter,
			Offset:         nextPageOffset,
			Limit:          &limit,
			WithPayload:    qdrant.NewWithPayload(true),
			WithVectors:    qdrant.NewWithVectors(false),
		})
		if err != nil {
			if strings.Contains(err.Error(), "doesn't exist") {
				collectionName = "knowledge_base"
				continue
			}
			return nil, fmt.Errorf("Scroll 查询失败: %w", err)
		}

		for _, point := range scrollResult {
			if len(results) >= maxResults {
				break
			}

			payload := make(map[string]interface{})
			if point.Payload != nil {
				for key, value := range point.Payload {
					payload[key] = scrollExtractQdrantValue(value)
				}
			}

			matched := lowerKeyword == ""
			if !matched {
				for _, value := range payload {
					if valueStr, ok := value.(string); ok && strings.Contains(strings.ToLower(valueStr), lowerKeyword) {
						matched = true
						break
					}
				}
			}

			if point.Id != nil {
				nextPageOffset = point.Id
			}
			if !matched {
				continue
			}

			var id interface{}
			if point.Id != nil {
				id = scrollExtractQdrantID(point.Id)
			}
			results = append(results, KeywordScrollResult{
				ID:      id,
				Score:   1.0,
				Payload: payload,
			})
		}

		if len(scrollResult) < int(limit) {
			break
		}
	}

	return results, nil
}

// ScrollPointsByFileName 按 metadata.file_name 精确匹配滚动拉取全部点
func (qc *QdrantTools) ScrollPointsByFileName(ctx context.Context, collectionName, fileName string) ([]QdrantScrollPoint, error) {
	if qc.client == nil {
		return nil, fmt.Errorf("Qdrant 客户端未初始化")
	}
	return scrollPointsByFileNameSDK(ctx, qc.client, collectionName, fileName)
}

// GeneratePointID 生成点的唯一 ID
func GeneratePointID(filePath string, page int) string {
	// 使用文件路径和页码生成唯一 ID
	hash := md5.Sum([]byte(fmt.Sprintf("%s_%d", filePath, page)))
	return fmt.Sprintf("%x", hash)
}

// GenerateUUID 生成 UUID
func GenerateUUID() string {
	return uuid.New().String()
}

// FormatCollectionName 根据前缀、租户ID、文档类型代码及是否是全文摘要集合，生成标准的 Qdrant collection 名称。
// 格式：{prefix}{clientId}_{docCode}[_全文]
// 如果 clientId 为空，则回退到旧格式：{prefix}{docCode}[_全文]
func FormatCollectionName(prefix, clientId, docCode string, isSummary bool) string {
	prefix = strings.TrimSpace(prefix)
	if prefix == "" {
		prefix = "raw_doc_"
	}
	clientId = strings.TrimSpace(clientId)
	docCode = strings.TrimSpace(docCode)

	suffix := ""
	if isSummary {
		suffix = "_全文"
	}

	if clientId == "" {
		return prefix + docCode + suffix
	}
	return fmt.Sprintf("%s%s_%s%s", prefix, clientId, docCode, suffix)
}

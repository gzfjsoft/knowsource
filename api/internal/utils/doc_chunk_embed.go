package utils

import (
	"context"
	"fmt"

	"github.com/zeromicro/go-zero/core/logx"
)

const (
	DefaultChunkSize    = 1000
	DefaultChunkOverlap = 100
)

// IndexDocStats 记录索引执行统计信息，便于上层将分块结果回传给调用方。
type IndexDocStats struct {
	RequestedMode ChunkMode `json:"requestedMode"`
	EffectiveMode ChunkMode `json:"effectiveMode"`
	ChunkCount    int       `json:"chunkCount"`
	ChunkSize     int       `json:"chunkSize"`
	ChunkOverlap  int       `json:"chunkOverlap"`
}

// DocMeta 文档元信息，与 document_indexer.py 中写入 Qdrant 的 metadata 对齐
type DocMeta struct {
	Path            string                 `json:"path"`              // 文件路径
	Page            int                    `json:"page"`              // 当前块页码（从 0 起）
	TotalPages      int                    `json:"total_pages"`       // 总块数
	Length          int                    `json:"length"`            // 当前块字符长度
	FileMD5         string                 `json:"file_md5"`          // 文件 MD5
	FileCreatedTime float64                `json:"file_created_time"` // 文件创建时间（Unix 时间戳）
	FileSize        int64                  `json:"file_size"`         // 文件大小（字节）
	DocType         string                 `json:"doc_type"`          // 文档类型
	Tag             string                 `json:"tag"`               // 标签
	FileName        string                 `json:"file_name"`         // 文件名
	Extra           map[string]interface{} `json:"-"`                 // 其它 meta（如从 .meta.yaml 来的字段），一并写入 payload
}

// ChunkDocContent 根据分块模式对文档内容分块，与 chunk_manager.py 行为一致
// 返回分块文本列表。若为 ai_separator 且内容不含 <AI分隔符> 则 fallback 使用 simple。
func ChunkDocContent(content string, mode ChunkMode, chunkSize, chunkOverlap int) []string {
	m := NewMarkDownChunk()
	chunkSize, chunkOverlap = normalizeChunkParams(chunkSize, chunkOverlap)
	switch mode {
	case ChunkModeNone:
		return m.ChunkNone(content)
	case ChunkModeAISeparator:
		chunks := m.ChunkByAISeparator(content)
		if len(chunks) > 0 {
			return chunks
		}
		// 无 <AI分隔符> 时退回 simple
		return m.ChunkSimple(content, chunkSize, chunkOverlap)
	case ChunkModeSmart:
		return m.ChunkSmart(content, chunkSize)
	case ChunkModeSuperSmart:
		chunks := m.ChunkSuperSmart(content, chunkSize)
		// SuperSmart 在极端 Markdown 输入下可能解析不到块，回退到 ChunkSmart 确保可写入 Qdrant
		if len(chunks) > 0 {
			return chunks
		}
		return m.ChunkSmart(content, chunkSize)
	case ChunkModeSimple:
		fallthrough
	default:
		return m.ChunkSimple(content, chunkSize, chunkOverlap)
	}
}

func normalizeChunkParams(chunkSize, chunkOverlap int) (int, int) {
	if chunkSize <= 0 {
		chunkSize = DefaultChunkSize
	}
	if chunkOverlap < 0 {
		chunkOverlap = 0
	}
	if chunkOverlap >= chunkSize {
		chunkOverlap = chunkSize / 5
	}
	return chunkSize, chunkOverlap
}

func resolveEffectiveMode(content string, mode ChunkMode, chunkSize, chunkOverlap int) ChunkMode {
	m := NewMarkDownChunk()
	chunkSize, chunkOverlap = normalizeChunkParams(chunkSize, chunkOverlap)
	switch mode {
	case ChunkModeNone:
		return ChunkModeNone
	case ChunkModeAISeparator:
		if len(m.ChunkByAISeparator(content)) > 0 {
			return ChunkModeAISeparator
		}
		return ChunkModeSimple
	case ChunkModeSuperSmart:
		if len(m.ChunkSuperSmart(content, chunkSize)) > 0 {
			return ChunkModeSuperSmart
		}
		return ChunkModeSimple
	case ChunkModeSmart:
		return ChunkModeSmart
	default:
		return ChunkModeSimple
	}
}

// IndexDocToQdrant 将文档内容分块、调用 Ollama 向量化并写入 Qdrant
// 与 document_indexer.py 的 add_documents_to_qdrant 对齐：payload["metadata"] 含 path/page/total_pages/length/file_md5/file_created_time/file_size 及 extra
// qdrantClient 需已设置 EmbeddingAPI 为 Ollama 地址（如 http://localhost:11434），将使用 /api/embeddings
func IndexDocToQdrant(ctx context.Context, content string, meta DocMeta, collectionName string, qdrantClient *QdrantTools, chunkMode ChunkMode, chunkSize, chunkOverlap int) error {
	_, err := IndexDocToQdrantWithStats(ctx, content, meta, collectionName, qdrantClient, chunkMode, chunkSize, chunkOverlap)
	return err
}

// IndexDocToQdrantWithStats 在原有写入逻辑上返回分块统计信息。
func IndexDocToQdrantWithStats(ctx context.Context, content string, meta DocMeta, collectionName string, qdrantClient *QdrantTools, chunkMode ChunkMode, chunkSize, chunkOverlap int) (*IndexDocStats, error) {
	if content == "" {
		return nil, fmt.Errorf("文档内容为空")
	}
	if qdrantClient == nil {
		return nil, fmt.Errorf("QdrantTools 为空")
	}
	if qdrantClient.EmbeddingAPI == "" && qdrantClient.VllmEmbeddingURL == "" {
		return nil, fmt.Errorf("未配置 Embedding API（Rag.EmbeddingsUrl 或 Ollama 地址）")
	}

	chunkSize, chunkOverlap = normalizeChunkParams(chunkSize, chunkOverlap)
	effectiveMode := resolveEffectiveMode(content, chunkMode, chunkSize, chunkOverlap)
	chunks := ChunkDocContent(content, chunkMode, chunkSize, chunkOverlap)
	if len(chunks) == 0 {
		return nil, fmt.Errorf("分块结果为空")
	}

	// 分块结果日志：块数、每块长度及前 80 字预览
	logx.WithContext(ctx).Infof("分块结果 path=%s requested=%s effective=%s chunkSize=%d overlap=%d 共 %d 块", meta.Path, chunkMode, effectiveMode, chunkSize, chunkOverlap, len(chunks))
	for idx, c := range chunks {
		preview := c
		if len(preview) > 80 {
			preview = preview[:80] + "..."
		}
		logx.WithContext(ctx).Infof("  块[%d] len=%d preview=%s", idx, len(c), preview)
	}

	// 确保集合存在并获取向量维度（用第一块请求一次 embedding 得到 size）
	firstVec, err := qdrantClient.GenerateEmbedding(ctx, chunks[0])
	if err != nil {
		return nil, fmt.Errorf("生成首块向量失败: %w", err)
	}
	vectorSize := len(firstVec)

	exists, err := qdrantClient.CollectionExists(ctx, collectionName)
	if err != nil {
		return nil, fmt.Errorf("检查集合是否存在: %w", err)
	}
	if !exists {
		if err := qdrantClient.CreateCollection(ctx, collectionName, vectorSize, "Dot"); err != nil {
			return nil, fmt.Errorf("创建集合: %w", err)
		}
		if err := qdrantClient.CreatePayloadIndex(ctx, collectionName, "metadata.path", "keyword"); err != nil {
			logx.WithContext(ctx).Errorf("创建 payload 索引失败: %v", err)
		}
	}

	const batchSize = 10
	for batchIdx := 0; batchIdx < (len(chunks)+batchSize-1)/batchSize; batchIdx++ {
		start := batchIdx * batchSize
		end := start + batchSize
		if end > len(chunks) {
			end = len(chunks)
		}
		batchChunks := chunks[start:end]
		points := make([]Point, 0, len(batchChunks))

		for i, chunk := range batchChunks {
			page := start + i
			var vector []float32
			if page == 0 {
				vector = firstVec // 复用已算的首块向量
			} else {
				var err error
				vector, err = qdrantClient.GenerateEmbedding(ctx, chunk)
				if err != nil {
					return nil, fmt.Errorf("生成第 %d 块向量失败: %w", page, err)
				}
			}

			// 与 document_indexer.py 一致：metadata 含 path, page, total_pages, length, file_md5, file_created_time, file_size，以及 doc_type/tag/file_name 和 extra
			md := map[string]interface{}{
				"path":              meta.Path,
				"page":              page,
				"total_pages":       len(chunks),
				"length":            len(chunk),
				"file_md5":          meta.FileMD5,
				"file_created_time": meta.FileCreatedTime,
				"file_size":         meta.FileSize,
			}
			if meta.DocType != "" {
				md["doc_type"] = meta.DocType
			}
			if meta.Tag != "" {
				md["tag"] = meta.Tag
			}
			if meta.FileName != "" {
				md["file_name"] = meta.FileName
			}
			for k, v := range meta.Extra {
				md[k] = v
			}

			pointID := GeneratePointID(meta.Path, page)
			payload := map[string]interface{}{
				"metadata":     md,
				"page_content": chunk, // 分块原文，便于检索后直接展示
			}
			points = append(points, Point{
				ID:      pointID,
				Vector:  vector,
				Payload: payload,
			})
		}

		if err := qdrantClient.UpsertPoints(ctx, collectionName, points); err != nil {
			return nil, fmt.Errorf("批量写入 Qdrant 失败: %w", err)
		}
		logx.WithContext(ctx).Infof("已写入批次 %d/%d (%d 个点)", batchIdx+1, (len(chunks)+batchSize-1)/batchSize, len(points))
	}

	logx.WithContext(ctx).Infof("文档索引完成 collection=%s path=%s chunks=%d", collectionName, meta.Path, len(chunks))
	return &IndexDocStats{
		RequestedMode: chunkMode,
		EffectiveMode: effectiveMode,
		ChunkCount:    len(chunks),
		ChunkSize:     chunkSize,
		ChunkOverlap:  chunkOverlap,
	}, nil
}

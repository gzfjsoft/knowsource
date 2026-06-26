package utils

import (
	"context"
	"crypto/md5"
	"fmt"
	"io"
	"knowsource/model"
	"os"

	"github.com/zeromicro/go-zero/core/logx"
)

// ChunkConfig 分块配置
type ChunkConfig struct {
	ChunkSize    int       // 分块大小
	ChunkOverlap int       // 重叠大小
	ChunkMode    ChunkMode // 分块模式
}

// DefaultChunkConfig 默认分块配置
func DefaultChunkConfig() *ChunkConfig {
	return &ChunkConfig{
		ChunkSize:    5000,
		ChunkOverlap: 200,
		ChunkMode:    ChunkModeSmart,
	}
}

// DocumentIndexer 文档索引服务
type DocumentIndexer struct {
	QdrantTools  *QdrantTools
	ChunkManager *MarkDownChunk
	ChunkConfig  *ChunkConfig
	VectorSize   int // 向量维度
}

// NewDocumentIndexer 创建新的文档索引服务
func NewDocumentIndexer(qdrantClient *QdrantTools, vectorSize int) *DocumentIndexer {
	return &DocumentIndexer{
		QdrantTools:  qdrantClient,
		ChunkManager: NewMarkDownChunk(),
		ChunkConfig:  DefaultChunkConfig(),
		VectorSize:   vectorSize,
	}
}

// SetChunkConfig 设置分块配置
func (di *DocumentIndexer) SetChunkConfig(config *ChunkConfig) {
	di.ChunkConfig = config
}

// IndexDocument 将文档索引到 Qdrant
func (di *DocumentIndexer) IndexDocument(ctx context.Context, rawDoc *model.RawDocuments, collectionName string) error {
	// 检查文档是否已转换为 MD
	if rawDoc.IsToMd != 1 || rawDoc.Content == "" {
		return fmt.Errorf("文档尚未转换为 Markdown，无法索引")
	}

	// 获取文件信息
	filePath := rawDoc.FilePath
	fileInfo, err := di.getFileInfo(filePath)
	if err != nil {
		return fmt.Errorf("获取文件信息失败: %v", err)
	}

	// 根据分块模式进行分块（与 chunk_manager.py 一致）
	var chunks []string
	switch di.ChunkConfig.ChunkMode {
	case ChunkModeNone:
		chunks = di.ChunkManager.ChunkNone(rawDoc.Content)
	case ChunkModeSimple:
		chunks = di.ChunkManager.ChunkSimple(rawDoc.Content, di.ChunkConfig.ChunkSize, di.ChunkConfig.ChunkOverlap)
	case ChunkModeSmart:
		chunks = di.ChunkManager.ChunkSmart(rawDoc.Content, di.ChunkConfig.ChunkSize)
	case ChunkModeSuperSmart:
		chunks = di.ChunkManager.ChunkSuperSmart(rawDoc.Content, di.ChunkConfig.ChunkSize)
	case ChunkModeAISeparator:
		chunks = di.ChunkManager.ChunkByAISeparator(rawDoc.Content)
		if len(chunks) == 0 {
			chunks = di.ChunkManager.ChunkSimple(rawDoc.Content, di.ChunkConfig.ChunkSize, di.ChunkConfig.ChunkOverlap)
		}
	default:
		chunks = di.ChunkManager.ChunkSimple(rawDoc.Content, di.ChunkConfig.ChunkSize, di.ChunkConfig.ChunkOverlap)
	}

	logx.WithContext(ctx).Infof("文档 %s 分块完成，共 %d 块", rawDoc.FileName, len(chunks))

	// 确保集合存在
	exists, err := di.QdrantTools.CollectionExists(ctx, collectionName)
	if err != nil {
		return fmt.Errorf("检查集合是否存在失败: %v", err)
	}

	if !exists {
		// 创建集合
		err = di.QdrantTools.CreateCollection(ctx, collectionName, di.VectorSize, "Dot")
		if err != nil {
			return fmt.Errorf("创建集合失败: %v", err)
		}

		// 创建 payload 索引
		err = di.QdrantTools.CreatePayloadIndex(ctx, collectionName, "metadata.path", "keyword")
		if err != nil {
			logx.WithContext(ctx).Errorf("创建 payload 索引失败: %v", err)
		}

		logx.WithContext(ctx).Infof("创建集合: %s", collectionName)
	}

	// 批量插入文档块
	batchSize := 10
	totalBatches := (len(chunks) + batchSize - 1) / batchSize

	for batchIdx := 0; batchIdx < totalBatches; batchIdx++ {
		startIdx := batchIdx * batchSize
		endIdx := startIdx + batchSize
		if endIdx > len(chunks) {
			endIdx = len(chunks)
		}

		batchChunks := chunks[startIdx:endIdx]
		batchPoints := make([]Point, 0, len(batchChunks))

		for i, chunk := range batchChunks {
			page := startIdx + i

			// 生成向量（需要实现 embedding 生成）
			vector, err := di.QdrantTools.GenerateEmbedding(ctx, chunk)
			if err != nil {
				logx.WithContext(ctx).Errorf("生成向量失败: %v", err)
				// 如果无法生成向量，使用零向量（实际应用中应该处理这个错误）
				vector = make([]float32, di.VectorSize)
			}

			// 生成点的 ID
			pointID := GeneratePointID(filePath, page)

			// 构建 metadata
			metadata := map[string]interface{}{
				"path":          filePath,
				"page":          page,
				"total_pages":   len(chunks),
				"length":        len(chunk),
				"file_md5":      rawDoc.FileMd5,
				"file_size":     rawDoc.FileSize,
				"file_name":     rawDoc.FileName,
				"document_code": rawDoc.DocumentCode,
				"rawdoc_id":     rawDoc.Id,
			}

			// 添加文件创建时间
			if fileInfo != nil {
				if fileInfo.CreatedTime > 0 {
					metadata["file_created_time"] = fileInfo.CreatedTime
				}
			}

			point := Point{
				ID:      pointID,
				Vector:  vector,
				Payload: map[string]interface{}{"metadata": metadata},
			}

			batchPoints = append(batchPoints, point)
		}

		// 批量插入
		err = di.QdrantTools.UpsertPoints(ctx, collectionName, batchPoints)
		if err != nil {
			return fmt.Errorf("批量插入失败: %v", err)
		}

		logx.WithContext(ctx).Infof("已插入批次 %d/%d (%d 个文档)", batchIdx+1, totalBatches, len(batchPoints))
	}

	logx.WithContext(ctx).Infof("文档 %s 索引完成，共 %d 个块", rawDoc.FileName, len(chunks))
	return nil
}

// FileInfo 文件信息
type FileInfo struct {
	Path         string
	Size         int64
	MD5          string
	CreatedTime  int64
	ModifiedTime int64
}

// getFileInfo 获取文件信息
func (di *DocumentIndexer) getFileInfo(filePath string) (*FileInfo, error) {
	// 如果 filePath 是相对路径，需要拼接完整路径
	// 这里假设 filePath 已经是完整路径或相对于某个根目录
	stat, err := os.Stat(filePath)
	if err != nil {
		return nil, err
	}

	fileInfo := &FileInfo{
		Path:         filePath,
		Size:         stat.Size(),
		CreatedTime:  stat.ModTime().Unix(),
		ModifiedTime: stat.ModTime().Unix(),
	}

	// 计算 MD5
	file, err := os.Open(filePath)
	if err != nil {
		return fileInfo, nil // 即使无法计算 MD5，也返回基本信息
	}
	defer file.Close()

	hash := md5.New()
	if _, err := io.Copy(hash, file); err == nil {
		fileInfo.MD5 = fmt.Sprintf("%x", hash.Sum(nil))
	}

	return fileInfo, nil
}

// BatchIndexDocuments 批量索引文档
func (di *DocumentIndexer) BatchIndexDocuments(ctx context.Context, rawDocs []*model.RawDocuments, collectionName string) (successCount int, failCount int, errors []error) {
	successCount = 0
	failCount = 0
	errors = []error{}

	for _, rawDoc := range rawDocs {
		err := di.IndexDocument(ctx, rawDoc, collectionName)
		if err != nil {
			failCount++
			errors = append(errors, fmt.Errorf("文档 %s (ID: %d) 索引失败: %v", rawDoc.FileName, rawDoc.Id, err))
			logx.WithContext(ctx).Errorf("文档索引失败: %v", err)
		} else {
			successCount++
			logx.WithContext(ctx).Infof("文档索引成功: %s (ID: %d)", rawDoc.FileName, rawDoc.Id)
		}
	}

	return successCount, failCount, errors
}

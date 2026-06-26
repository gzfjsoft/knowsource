package ai

import (
	"context"
	"fmt"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	hdLogic "knowsource/api/internal/logic/knowdata"
	knowsourceLogic "knowsource/api/internal/logic/knowsource"
	"knowsource/api/internal/svc"
	"knowsource/api/internal/types"
	"knowsource/api/internal/utils"

	"github.com/zeromicro/go-zero/core/logx"
)

const (
	RagSourceFulltext = "fulltext"
	RagSourceSummary  = "summary"
	RagSourceVector   = "vector"
)

type RagRetrieveRequest struct {
	ClientId        string
	Message         string
	DocumentCode    string
	Tags            []string
	SkipRag         bool
	HasUploadedDocs bool
}

type RagRawDocMeta struct {
	Id       int64
	FileName string
}

// RagDocument 单条 RAG 命中（全文 / 概要向量 / 子块向量），与聊天注入用的 ragContext 分离
type RagDocument struct {
	Source          string
	Content         string
	FileName        string
	Path            string
	RawDocId        int64
	Page            int
	QdrantId        string
	SimilarityScore float64
	RerankScore     float64
}

type RagRetrieveResult struct {
	Documents        []RagDocument
	RagContext       string // 供会话聊天注入 LLM：仅由 fulltext + vector 子块拼成，不含 summary 行
	FileInfos        []string
	FileSimilarities []string
	RawDocs          []RagRawDocMeta
	FullTextMs       int64
	MainSearchMs     int64
	SubSearchMs      int64
	RawSearchUsed    bool
	VectorSearchUsed bool
	RerankConfigured bool // cfg.Rag.RerankerUrl 是否非空
	SummaryRerank    *types.SearchRawDocVectorsRerankInfo
	ChunkRerank      *types.SearchRawDocVectorsRerankInfo
}

type RagRetrievalOrchestrator struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewRagRetrievalOrchestrator(ctx context.Context, svcCtx *svc.ServiceContext) *RagRetrievalOrchestrator {
	return &RagRetrievalOrchestrator{
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (o *RagRetrievalOrchestrator) Retrieve(req RagRetrieveRequest) (*RagRetrieveResult, error) {
	result := &RagRetrieveResult{
		RerankConfigured: strings.TrimSpace(o.svcCtx.Config.Rag.RerankerUrl) != "",
	}
	if req.SkipRag {
		return result, nil
	}

	collectionName := strings.TrimSpace(req.DocumentCode)
	if collectionName == "" || req.HasUploadedDocs {
		return result, nil
	}

	ftReq := &types.SearchRawDocumentsRequest{
		Keyword:      req.Message,
		DocumentCode: req.DocumentCode,
		Tag:          strings.Join(req.Tags, ","),
		PageRequest:  types.PageRequest{Page: 1, PageSize: 10},
	}
	ftLogic := hdLogic.NewSearchRawDocumentsLogic(o.ctx, o.svcCtx)
	ftStart := time.Now()
	ftResp, ftErr := ftLogic.SearchRawDocuments(ftReq)
	result.FullTextMs = time.Since(ftStart).Milliseconds()
	if ftErr != nil {
		logx.Errorf("RAG 全文检索 raw-documents/search 调用失败: %v", ftErr)
	} else if ftResp != nil && ftResp.Code == 200 && len(ftResp.Data.List) > 0 {
		for _, item := range ftResp.Data.List {
			if item.FileName != "" {
				result.FileInfos = append(result.FileInfos, item.FileName)
			}
			if item.Id != 0 {
				result.RawDocs = append(result.RawDocs, RagRawDocMeta{Id: item.Id, FileName: item.FileName})
			}
			snippet := strings.TrimSpace(item.Snippet)
			if snippet == "" {
				continue
			}
			result.Documents = append(result.Documents, RagDocument{
				Source:          RagSourceFulltext,
				Content:         snippet,
				FileName:        item.FileName,
				RawDocId:        item.Id,
				SimilarityScore: 1,
				RerankScore:     0,
			})
		}
		result.RagContext = buildRagContextFromDocuments(result.Documents)
		result.RawSearchUsed = true
		logx.Infof("RAG 全文检索命中 %d 条，已注入共 %d 字符，跳过向量检索", len(ftResp.Data.List), len(result.RagContext))
	}

	if result.RagContext != "" {
		return result, nil
	}

	prefix := o.svcCtx.Config.Qdrant.CollectionPrefix
	fullTextCollection := utils.FormatCollectionName(prefix, req.ClientId, req.DocumentCode, true)
	ragTopK, ragRerankK := 10, 5
	ragSimThreshold, ragScoreThreshold := 0.0, 0.0
	summarySimThreshold, summaryScoreThreshold := 0.0, 0.0
	if llmCfg, err := knowsourceLogic.LoadAIConfig(req.ClientId); err == nil {
		if llmCfg.RagEmbeddingTopK > 0 {
			ragTopK = int(llmCfg.RagEmbeddingTopK)
		}
		if llmCfg.RagRerankTopK > 0 {
			ragRerankK = int(llmCfg.RagRerankTopK)
		}
		ragSimThreshold = llmCfg.RagSimilarityThreshold
		ragScoreThreshold = llmCfg.RagRerankScoreThreshold
		// 概要检索阈值优先使用单独配置；未配置时 LoadAIConfig 会做兼容回落
		summarySimThreshold = llmCfg.RagSummarySimilarityThreshold
		summaryScoreThreshold = llmCfg.RagSummaryRerankScoreThreshold
	}

	var queryVector []float32
	qc, qErr := utils.NewQdrantToolsWithEmbeddingForClient(&o.svcCtx.Config, req.ClientId)
	if qErr != nil {
		logx.Errorf("初始化 QdrantTools 失败: %v", qErr)
	} else {
		queryVector, qErr = qc.GenerateEmbedding(o.ctx, req.Message)
		if qErr != nil {
			logx.Errorf("生成查询向量失败: %v", qErr)
		}
	}

	vecLogic := hdLogic.NewSearchRawDocVectorsLogic(o.ctx, o.svcCtx)
	fullTextReq := &types.SearchRawDocVectorsRequest{
		Query:                req.Message,
		TopK:                 2,
		UseRerank:            true,
		RerankTopK:           2,
		SimilarityThreshold:  summarySimThreshold,
		RerankScoreThreshold: summaryScoreThreshold,
		CollectionName:       fullTextCollection,
		Tags:                 req.Tags,
		QueryVector:          queryVector,
	}
	mainStart := time.Now()
	fullTextResp, fullTextErr := vecLogic.SearchRawDocVectors(fullTextReq)
	result.MainSearchMs = time.Since(mainStart).Milliseconds()
	if fullTextErr == nil && fullTextResp != nil && fullTextResp.Code == 200 && fullTextResp.Data != nil && fullTextResp.Data.RerankInfo != nil {
		ri := *fullTextResp.Data.RerankInfo
		result.SummaryRerank = &ri
	}

	var filenames []string
	if fullTextErr == nil && fullTextResp != nil && fullTextResp.Code == 200 && fullTextResp.Data != nil && len(fullTextResp.Data.List) > 0 {
		for _, item := range fullTextResp.Data.List {
			fileName := fileNameFromItem(item)
			if fileName == "" {
				continue
			}
			filenames = append(filenames, fileName)
			result.FileSimilarities = append(result.FileSimilarities, fmt.Sprintf("%s (相似度: %.4f)", fileName, item.SimilarityScore))
			content := strings.TrimSpace(item.Content)
			if content != "" {
				result.Documents = append(result.Documents, RagDocument{
					Source:          RagSourceSummary,
					Content:         content,
					FileName:        fileName,
					Path:            item.Path,
					RawDocId:        parseRawDocID(item.Metadata),
					Page:            pageFromMeta(item.Metadata),
					QdrantId:        item.QdrantId,
					SimilarityScore: item.SimilarityScore,
					RerankScore:     item.RerankScore,
				})
			}
		}
	}

	mainCollection := utils.FormatCollectionName(prefix, req.ClientId, req.DocumentCode, false)
	vecReq := &types.SearchRawDocVectorsRequest{
		Query:                req.Message,
		TopK:                 ragTopK,
		UseRerank:            true,
		RerankTopK:           ragRerankK,
		SimilarityThreshold:  ragSimThreshold,
		RerankScoreThreshold: ragScoreThreshold,
		CollectionName:       mainCollection,
		Tags:                 req.Tags,
		QueryVector:          queryVector,
		FileNames:            filenames,
	}
	subStart := time.Now()
	vecResp, vecErr := vecLogic.SearchRawDocVectors(vecReq)
	result.SubSearchMs = time.Since(subStart).Milliseconds()
	if vecErr == nil && vecResp != nil && vecResp.Code == 200 && vecResp.Data != nil && vecResp.Data.RerankInfo != nil {
		ri := *vecResp.Data.RerankInfo
		result.ChunkRerank = &ri
	}
	if vecErr != nil {
		logx.Errorf("RAG vectors/search 调用失败: %v", vecErr)
		return result, nil
	}
	if vecResp == nil || vecResp.Code != 200 || vecResp.Data == nil || len(vecResp.Data.List) == 0 {
		return result, nil
	}

	sortVectorList(vecResp.Data.List)
	for _, item := range vecResp.Data.List {
		if strings.TrimSpace(item.Content) == "" {
			continue
		}
		rawdocID := parseRawDocID(item.Metadata)
		fileName := fileNameFromItem(item)
		result.Documents = append(result.Documents, RagDocument{
			Source:          RagSourceVector,
			Content:         item.Content,
			FileName:        fileName,
			Path:            item.Path,
			RawDocId:        rawdocID,
			Page:            pageFromMeta(item.Metadata),
			QdrantId:        item.QdrantId,
			SimilarityScore: item.SimilarityScore,
			RerankScore:     item.RerankScore,
		})
		if fileName != "" {
			result.FileInfos = append(result.FileInfos, fileName)
		}
		if rawdocID != 0 {
			result.RawDocs = append(result.RawDocs, RagRawDocMeta{Id: rawdocID, FileName: fileName})
		}
	}
	result.RagContext = buildRagContextFromDocuments(result.Documents)
	result.VectorSearchUsed = hasVectorChunkContent(result.Documents)
	logx.Infof("RAG 向量检索已注入 %d 条，共 %d 字符", len(vecResp.Data.List), len(result.RagContext))
	return result, nil
}

func hasVectorChunkContent(docs []RagDocument) bool {
	for _, d := range docs {
		if d.Source == RagSourceVector && strings.TrimSpace(d.Content) != "" {
			return true
		}
	}
	return false
}

// buildRagContextFromDocuments 拼 LLM 上下文：仅 fulltext + vector 子块，不包含 summary 概要点
func buildRagContextFromDocuments(docs []RagDocument) string {
	var sb strings.Builder
	first := true
	for _, d := range docs {
		if d.Source == RagSourceSummary {
			continue
		}
		if strings.TrimSpace(d.Content) == "" {
			continue
		}
		if !first {
			sb.WriteString("\n\n")
		}
		first = false
		if d.FileName != "" {
			sb.WriteString(fmt.Sprintf("<filename>%s</filename>\n", d.FileName))
		}
		if d.Source == RagSourceVector {
			sb.WriteString(fmt.Sprintf("<similarity>%.4f</similarity>\n", d.SimilarityScore))
			if d.RerankScore > 0 {
				sb.WriteString(fmt.Sprintf("<rerank_score>%.4f</rerank_score>\n", d.RerankScore))
			}
		}
		sb.WriteString(d.Content)
	}
	return sb.String()
}

func sortVectorList(list []types.SearchRawDocVectorsItem) {
	sort.Slice(list, func(i, j int) bool {
		if list[i].RerankScore != list[j].RerankScore {
			return list[i].RerankScore > list[j].RerankScore
		}
		fileNameI := fileNameFromItem(list[i])
		fileNameJ := fileNameFromItem(list[j])
		if fileNameI != fileNameJ {
			return fileNameI < fileNameJ
		}
		return pageFromMeta(list[i].Metadata) < pageFromMeta(list[j].Metadata)
	})
}

func fileNameFromItem(item types.SearchRawDocVectorsItem) string {
	if item.Path != "" {
		return filepath.Base(item.Path)
	}
	if item.Metadata != nil {
		if fn, ok := item.Metadata["file_name"].(string); ok {
			return fn
		}
	}
	return ""
}

func parseRawDocID(meta map[string]interface{}) int64 {
	if meta == nil {
		return 0
	}
	v, ok := meta["rawdoc_id"]
	if !ok {
		return 0
	}
	switch vv := v.(type) {
	case float64:
		return int64(vv)
	case int:
		return int64(vv)
	case int64:
		return vv
	case string:
		id, err := strconv.ParseInt(vv, 10, 64)
		if err == nil {
			return id
		}
	}
	return 0
}

func pageFromMeta(meta map[string]interface{}) int {
	if meta == nil {
		return 0
	}
	v, ok := meta["page"]
	if !ok {
		return 0
	}
	switch vv := v.(type) {
	case float64:
		return int(vv)
	case int:
		return vv
	case int64:
		return int(vv)
	case string:
		p, err := strconv.Atoi(vv)
		if err == nil {
			return p
		}
	}
	return 0
}

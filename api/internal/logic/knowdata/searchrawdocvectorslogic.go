// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package knowdata

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"knowsource/api/internal/svc"
	"knowsource/api/internal/types"
	"knowsource/api/internal/utils"
	"knowsource/common/response"

	"github.com/zeromicro/go-zero/core/logx"
)

type SearchRawDocVectorsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 原始文档向量检索（可选 vLLM 重排）
func NewSearchRawDocVectorsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SearchRawDocVectorsLogic {
	return &SearchRawDocVectorsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *SearchRawDocVectorsLogic) SearchRawDocVectors(req *types.SearchRawDocVectorsRequest) (resp *types.SearchRawDocVectorsResponse, err error) {
	cfg := l.svcCtx.Config
	if strings.TrimSpace(req.Query) == "" && len(req.QueryVector) == 0 {
		return &types.SearchRawDocVectorsResponse{
			Response: types.Response{Code: response.ServerErrorCode, Message: "查询内容不能为空"},
		}, nil
	}
	if strings.TrimSpace(req.CollectionName) == "" {
		return &types.SearchRawDocVectorsResponse{
			Response: types.Response{Code: response.ServerErrorCode, Message: "必须指定 collectionName"},
		}, nil
	}
	if cfg.Qdrant.Host == "" || cfg.Qdrant.Port <= 0 {
		return &types.SearchRawDocVectorsResponse{
			Response: types.Response{Code: response.ServerErrorCode, Message: "未配置 Qdrant 地址"},
		}, nil
	}
	if l.svcCtx.QdrantClient == nil {
		return &types.SearchRawDocVectorsResponse{
			Response: types.Response{Code: response.ServerErrorCode, Message: "Qdrant 客户端未初始化"},
		}, nil
	}

	topK := req.TopK
	if topK <= 0 {
		topK = 10
	}
	rerankTopK := req.RerankTopK
	if rerankTopK <= 0 {
		rerankTopK = topK
	}

	start := time.Now()
	queryVector := req.QueryVector
	if len(queryVector) == 0 {
		clientId, _ := l.ctx.Value("clientId").(string)
		qc, qErr := utils.NewQdrantToolsWithEmbeddingForClient(&cfg, clientId)
		if qErr != nil {
			return &types.SearchRawDocVectorsResponse{
				Response: types.Response{Code: response.ServerErrorCode, Message: qErr.Error()},
			}, nil
		}
		logx.Infof("生成查询向量 in searchrawdocvectorslogic: %s", req.Query)
		var err error
		queryVector, err = qc.GenerateEmbedding(l.ctx, req.Query)
		if err != nil {
			l.Errorf("生成查询向量失败: %v", err)
			return &types.SearchRawDocVectorsResponse{
				Response: types.Response{Code: response.ServerErrorCode, Message: "生成查询向量失败", Info: err.Error()},
			}, nil
		}
	}

	var filterTags []string
	for _, t := range req.Tags {
		if s := strings.TrimSpace(t); s != "" {
			filterTags = append(filterTags, s)
		}
	}

	// 使用 QdrantTools 检索
	var points []utils.ScoredPoint
	qt := l.svcCtx.QdrantTools
	if qt == nil {
		qt = utils.NewQdrantToolsWithClient(l.svcCtx.QdrantClient)
	}
	points, err = qt.SearchPointsWithFileFilter(l.ctx, req.CollectionName, queryVector, topK, filterTags, req.FileNames)
	if err != nil {
		l.Errorf("Qdrant 检索失败: %v", err)
		return &types.SearchRawDocVectorsResponse{
			Response: types.Response{Code: response.ServerErrorCode, Message: "检索失败", Info: err.Error()},
		}, nil
	}

	similarityThreshold := req.SimilarityThreshold
	list := make([]types.SearchRawDocVectorsItem, 0, len(points))
	for _, p := range points {
		if similarityThreshold > 0 && p.Score < similarityThreshold {
			continue
		}
		content, path, metadata := payloadContentAndMeta(p.Payload)
		if content == "" {
			continue
		}
		list = append(list, types.SearchRawDocVectorsItem{
			Id:              len(list),
			QdrantId:        idToString(p.ID),
			Path:            path,
			Content:         content,
			SimilarityScore: p.Score,
			Metadata:        metadata,
		})
	}

	rerankInfo := &types.SearchRawDocVectorsRerankInfo{UsedRerank: false}
	clientId, _ := l.ctx.Value("clientId").(string)
	rerankBaseURL, rerankType, rerankApiKey, rerankModel := utils.ResolveRerankRuntime(&cfg, clientId)
	if rerankType == "" {
		rerankType = "vllm"
	}
	if req.UseRerank && rerankBaseURL == "" && len(list) > 0 {
		rerankInfo.Error = "未配置 Rag.RerankerUrl，跳过重排（RerankScore 将保持为 0）"
	}
	if req.UseRerank && rerankBaseURL != "" && len(list) > 0 {
		documents := make([]string, len(list))
		for i := range list {
			documents[i] = list[i].Content
		}
		rerankResults, rerr := utils.RerankByTypeWithAPIKey(l.ctx, rerankBaseURL, rerankApiKey, rerankType, utils.RerankRequest{
			Query:     req.Query,
			Documents: documents,
			Model:     rerankModel,
		})
		if rerr != nil {
			l.Errorf("vLLM 重排失败: %v", rerr)
			rerankInfo.Error = rerr.Error()
		} else {
			scoreByIndex := make(map[int]float64)
			for _, r := range rerankResults {
				scoreByIndex[r.Index] = r.RelevanceScore
			}
			for i := range list {
				list[i].RerankScore = scoreByIndex[i]
			}
			// 按 rerank_score 降序，再截断到 rerankTopK，并过滤低于分值底线的
			sortByRerankScore(list)
			rerankScoreThreshold := req.RerankScoreThreshold
			if rerankScoreThreshold > 0 {
				filtered := list[:0]
				for i := range list {
					if list[i].RerankScore >= rerankScoreThreshold {
						filtered = append(filtered, list[i])
					}
				}
				list = filtered
			}
			if len(list) > rerankTopK {
				list = list[:rerankTopK]
			}
			rerankInfo.UsedRerank = true
			rerankInfo.RerankTopK = rerankTopK
			rerankInfo.OriginalCount = len(documents)
			rerankInfo.RerankedCount = len(list)
		}
	}

	searchTime := time.Since(start).Seconds()
	return &types.SearchRawDocVectorsResponse{
		Response: types.Response{Code: 200, Message: "success"},
		Data: &types.SearchRawDocVectorsData{
			Query:          req.Query,
			List:           list,
			Total:          len(list),
			SearchTime:     searchTime,
			TopK:           topK,
			CollectionName: req.CollectionName,
			RerankInfo:     rerankInfo,
		},
	}, nil
}

func payloadContentAndMeta(payload map[string]interface{}) (content, path string, metadata map[string]interface{}) {
	if payload == nil {
		return "", "", nil
	}
	if v, _ := payload["page_content"].(string); v != "" {
		content = v
	} else if v, _ := payload["content"].(string); v != "" {
		content = v
	} else if v, _ := payload["text"].(string); v != "" {
		content = v
	}
	if meta, _ := payload["metadata"].(map[string]interface{}); meta != nil {
		metadata = meta
		if p, _ := meta["path"].(string); p != "" {
			path = p
		}
	}
	return content, path, metadata
}

func idToString(id interface{}) string {
	if id == nil {
		return ""
	}
	return fmt.Sprint(id)
}

// sortByRerankScore 按 RerankScore 降序；无 rerank 时按 SimilarityScore 降序
func sortByRerankScore(list []types.SearchRawDocVectorsItem) {
	sort.Slice(list, func(i, j int) bool {
		si, sj := list[i].RerankScore, list[j].RerankScore
		if si == 0 && sj == 0 {
			si, sj = list[i].SimilarityScore, list[j].SimilarityScore
		}
		return si > sj
	})
}

// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package knowdata

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"knowsource/api/internal/svc"
	"knowsource/api/internal/types"
	"knowsource/api/internal/utils"
	"knowsource/common/response"
	"knowsource/model"

	"github.com/pkg/errors"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type GetRawDocumentQdrantChunksLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 查看已审核文档在 Qdrant 中的分块（主分块集合 + 全文概要集合，按 metadata.file_name 过滤）
func NewGetRawDocumentQdrantChunksLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetRawDocumentQdrantChunksLogic {
	return &GetRawDocumentQdrantChunksLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetRawDocumentQdrantChunksLogic) GetRawDocumentQdrantChunks(req *types.GetRawDocumentQdrantChunksRequest) (resp *types.GetRawDocumentQdrantChunksResponse, err error) {
	newResp := func(code int64, message, info string) *types.GetRawDocumentQdrantChunksResponse {
		return &types.GetRawDocumentQdrantChunksResponse{
			Response: types.Response{
				Code:    code,
				Message: message,
				Info:    info,
			},
		}
	}

	if req.Id <= 0 {
		return newResp(response.ParameterErrorCode, "文档 ID 无效", ""), nil
	}

	clientId, _ := l.ctx.Value("clientId").(string)
	clientId = strings.TrimSpace(clientId)
	if clientId == "" {
		return newResp(response.UnauthorizedCode, "clientId不能为空，请重新登录", ""), nil
	}

	doc, err := l.svcCtx.RawDocumentsModel.FindOneByClientId(l.ctx, clientId, req.Id)
	if err != nil {
		if err == sqlx.ErrNotFound || errors.Is(err, model.ErrNotFound) {
			return newResp(response.RecordNotExistCode, "文档不存在", ""), nil
		}
		l.Errorf("查询原始文档失败: %v", err)
		return newResp(response.ServerErrorCode, "查询失败", err.Error()), nil
	}

	if doc.IsAudit != 1 {
		return newResp(response.ParameterErrorCode, "仅已审核的文档可查看 Qdrant 分块", ""), nil
	}

	if strings.TrimSpace(doc.FileName) == "" {
		return newResp(response.ParameterErrorCode, "文档文件名为空，无法查询 Qdrant", ""), nil
	}

	cfg := l.svcCtx.Config
	if cfg.Qdrant.Host == "" || cfg.Qdrant.Port <= 0 {
		return newResp(response.ServerErrorCode, "未配置 Qdrant 地址", ""), nil
	}
	if l.svcCtx.QdrantClient == nil {
		return newResp(response.ServerErrorCode, "Qdrant 客户端未初始化", ""), nil
	}

	prefix := cfg.Qdrant.CollectionPrefix
	mainCollection := utils.FormatCollectionName(prefix, clientId, doc.DocumentCode, false)
	summaryCollection := utils.FormatCollectionName(prefix, clientId, doc.DocumentCode, true)

	qt := l.svcCtx.QdrantTools
	if qt == nil {
		qt = utils.NewQdrantToolsWithClient(l.svcCtx.QdrantClient)
	}
	mainPts, err := qt.ScrollPointsByFileName(l.ctx, mainCollection, doc.FileName)
	if err != nil {
		l.Errorf("scroll 主集合失败: %v", err)
		return newResp(response.ServerErrorCode, "查询 Qdrant 主集合失败", err.Error()), nil
	}

	sumPts, err := qt.ScrollPointsByFileName(l.ctx, summaryCollection, doc.FileName)
	if err != nil {
		l.Errorf("scroll 概要集合失败: %v", err)
		return newResp(response.ServerErrorCode, "查询 Qdrant 概要集合失败", err.Error()), nil
	}

	mainChunks := scrollPointsToChunkItems(mainPts, false)
	summaryChunks := scrollPointsToChunkItems(sumPts, true)

	sort.Slice(mainChunks, func(i, j int) bool {
		if mainChunks[i].PageIndex != mainChunks[j].PageIndex {
			return mainChunks[i].PageIndex < mainChunks[j].PageIndex
		}
		return mainChunks[i].QdrantId < mainChunks[j].QdrantId
	})
	sort.Slice(summaryChunks, func(i, j int) bool {
		if summaryChunks[i].PageIndex != summaryChunks[j].PageIndex {
			return summaryChunks[i].PageIndex < summaryChunks[j].PageIndex
		}
		return summaryChunks[i].QdrantId < summaryChunks[j].QdrantId
	})

	return &types.GetRawDocumentQdrantChunksResponse{
		Response: types.Response{
			Code:    response.SuccessCode,
			Message: "success",
		},
		Data: &types.GetRawDocumentQdrantChunksData{
			DocumentId:        doc.Id,
			FileName:          doc.FileName,
			DocumentCode:      doc.DocumentCode,
			IsAudit:           doc.IsAudit,
			MainCollection:    mainCollection,
			SummaryCollection: summaryCollection,
			MainChunks:        mainChunks,
			SummaryChunks:     summaryChunks,
			MainTotal:         int64(len(mainChunks)),
			SummaryTotal:      int64(len(summaryChunks)),
		},
	}, nil
}

const maxChunkContentRunes = 12000
const chunkPreviewRunes = 280

func scrollPointsToChunkItems(points []utils.QdrantScrollPoint, defaultSummary bool) []types.RawDocQdrantChunkItem {
	out := make([]types.RawDocQdrantChunkItem, 0, len(points))
	for _, p := range points {
		out = append(out, qdrantPointToChunkItem(p, defaultSummary))
	}
	return out
}

func qdrantPointToChunkItem(p utils.QdrantScrollPoint, defaultSummary bool) types.RawDocQdrantChunkItem {
	item := types.RawDocQdrantChunkItem{
		QdrantId: formatQdrantScrollID(p.ID),
	}

	pageContent := ""
	if p.Payload != nil {
		if v, ok := p.Payload["page_content"]; ok && v != nil {
			if s, ok2 := v.(string); ok2 {
				pageContent = s
			}
		}
	}

	if p.Payload != nil {
		if md, ok := p.Payload["metadata"].(map[string]interface{}); ok && md != nil {
			item.PageIndex = metaInt64(md, "page")
			item.TotalPages = metaInt64(md, "total_pages")
			item.Length = metaInt64(md, "length")
			item.Path = metaString(md, "path")
			item.IsSummary = metaBoolTrue(md, "is_summary") || defaultSummary
		}
	} else if defaultSummary {
		item.IsSummary = true
	}

	full := truncateRunes(pageContent, maxChunkContentRunes)
	item.Content = full
	item.ContentPreview = truncateRunes(full, chunkPreviewRunes)

	if item.Length == 0 && pageContent != "" {
		item.Length = int64(len([]rune(pageContent)))
	}

	return item
}

func formatQdrantScrollID(id interface{}) string {
	if id == nil {
		return ""
	}
	switch v := id.(type) {
	case string:
		return v
	case uint64:
		return fmt.Sprintf("%d", v)
	case int64:
		return fmt.Sprintf("%d", v)
	case int:
		return fmt.Sprintf("%d", v)
	case float64:
		return fmt.Sprintf("%.0f", v)
	case json.Number:
		return v.String()
	case map[string]interface{}:
		if u, ok := v["uuid"].(string); ok {
			return u
		}
		b, _ := json.Marshal(v)
		return string(b)
	default:
		return fmt.Sprintf("%v", v)
	}
}

func metaInt64(m map[string]interface{}, key string) int64 {
	v, ok := m[key]
	if !ok || v == nil {
		return 0
	}
	switch t := v.(type) {
	case float64:
		return int64(t)
	case int64:
		return t
	case int:
		return int64(t)
	default:
		return 0
	}
}

func metaString(m map[string]interface{}, key string) string {
	v, ok := m[key]
	if !ok || v == nil {
		return ""
	}
	if s, ok := v.(string); ok {
		return s
	}
	return fmt.Sprint(v)
}

func metaBoolTrue(m map[string]interface{}, key string) bool {
	v, ok := m[key]
	if !ok || v == nil {
		return false
	}
	if b, ok := v.(bool); ok {
		return b
	}
	if f, ok := v.(float64); ok {
		return f != 0
	}
	if s, ok := v.(string); ok {
		return strings.EqualFold(s, "true") || s == "1"
	}
	return false
}

func truncateRunes(s string, max int) string {
	if max <= 0 {
		return ""
	}
	r := []rune(s)
	if len(r) <= max {
		return s
	}
	return string(r[:max]) + "…"
}

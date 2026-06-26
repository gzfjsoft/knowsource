// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package knowdata

import (
	"net/http"

	"knowsource/api/internal/logic/knowdata"
	"knowsource/api/internal/svc"
	"knowsource/api/internal/types"

	"github.com/zeromicro/go-zero/rest/httpx"
)

// 查看已审核文档在 Qdrant 中的分块（主分块集合 + 全文概要集合，按 metadata.file_name 过滤）
func GetRawDocumentQdrantChunksHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.GetRawDocumentQdrantChunksRequest
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := knowdata.NewGetRawDocumentQdrantChunksLogic(r.Context(), svcCtx)
		resp, err := l.GetRawDocumentQdrantChunks(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}

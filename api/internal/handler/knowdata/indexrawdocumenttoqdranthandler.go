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

// 将原始文档分块并向量化写入 Qdrant
func IndexRawDocumentToQdrantHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.IndexRawDocumentToQdrantRequest
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := knowdata.NewIndexRawDocumentToQdrantLogic(r.Context(), svcCtx)
		resp, err := l.IndexRawDocumentToQdrant(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}

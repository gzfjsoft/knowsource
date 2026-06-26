// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package knowdata

import (
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"
	"knowsource/api/internal/logic/knowdata"
	"knowsource/api/internal/svc"
	"knowsource/api/internal/types"
)

// 查看审核入库时抽取的问答队列
func ListRawDocumentQaPairsHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.ListRawDocumentQaPairsRequest
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := knowdata.NewListRawDocumentQaPairsLogic(r.Context(), svcCtx)
		resp, err := l.ListRawDocumentQaPairs(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}

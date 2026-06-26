package knowdata

import (
	"net/http"

	"knowsource/api/internal/logic/knowdata"
	"knowsource/api/internal/svc"
	"knowsource/api/internal/types"

	"github.com/zeromicro/go-zero/rest/httpx"
)

// 获取原始文档的原始内容和当前内容用于比较
func GetRawDocumentsContentDiffHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.GetRawDocumentsContentDiffRequest
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := knowdata.NewGetRawDocumentsContentDiffLogic(r.Context(), svcCtx)
		resp, err := l.GetRawDocumentsContentDiff(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}

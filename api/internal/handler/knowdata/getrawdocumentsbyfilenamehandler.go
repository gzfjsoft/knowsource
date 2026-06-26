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

// 根据文件名获取原始文档（返回 Markdown 内容）
func GetRawDocumentsByFilenameHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.GetRawDocumentsByFilenameRequest
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := knowdata.NewGetRawDocumentsByFilenameLogic(r.Context(), svcCtx)
		resp, err := l.GetRawDocumentsByFilename(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}

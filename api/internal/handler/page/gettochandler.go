package page

import (
	"net/http"

	"knowsource/api/internal/logic/page"
	"knowsource/api/internal/svc"
	"knowsource/common/response"

	"github.com/zeromicro/go-zero/rest/httpx"
)

// Get TOC
func GetTocHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l := page.NewGetTocLogic(r.Context(), svcCtx)
		resp, err := l.GetToc()
		if err != nil {
			httpx.OkJsonCtx(r.Context(), w, response.FailWithInfo(response.InvalidRequestParamCodeInHandler, "Error while handle call", err.Error()))
		} else {
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			w.Write([]byte(resp.Body))

		}
	}
}

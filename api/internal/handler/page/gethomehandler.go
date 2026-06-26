package page

import (
	"net/http"

	"knowsource/api/internal/svc"
)

// Get home page
func GetHomeHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		// 重定向到index.html
		http.Redirect(w, r, "/static/index.html", http.StatusSeeOther)

		// l := page.NewGetHomeLogic(r.Context(), svcCtx)
		// resp, err := l.GetHome()

		// if err != nil {
		// 	httpx.OkJsonCtx(r.Context(), w, response.FailWithInfo(response.InvalidRequestParamCodeInHandler, "Error while handle call", err.Error()))
		// } else {
		// 	w.Header().Set("Content-Type", "text/html; charset=utf-8")
		// 	w.Write([]byte(resp.Body))

		// 	// httpx.OkJsonCtx(r.Context(), w, resp)
		// }
	}
}

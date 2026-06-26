package knowsource

import (
	"net/http"

	"knowsource/api/internal/logic/knowsource"
	"knowsource/api/internal/svc"

	"github.com/zeromicro/go-zero/rest/httpx"
)

// sync mysql
func SyncMysqlHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l := knowsource.NewSyncMysqlLogic(r.Context(), svcCtx)
		resp, err := l.SyncMysql()
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}

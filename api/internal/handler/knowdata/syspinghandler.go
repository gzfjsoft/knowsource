package knowdata

import (
	"net/http"

	"knowsource/api/internal/logic/knowdata"
	"knowsource/api/internal/svc"
	"knowsource/api/internal/utils"
)

func SysPingHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l := knowdata.NewSysPingLogic(r.Context(), svcCtx)
		resp, err := l.SysPing()
		utils.WriteResponse(w, r, err, "ping", resp)
	}
}

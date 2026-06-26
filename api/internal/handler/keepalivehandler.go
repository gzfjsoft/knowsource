package handler

import (
	"net/http"

	"knowsource/api/internal/logic"
	"knowsource/api/internal/svc"
	"knowsource/common/response"

	"github.com/zeromicro/go-zero/rest/httpx"
)

func KeepAliveHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l := logic.NewKeepAliveLogic(r.Context(), svcCtx)
		resp, err := l.KeepAlive()
		if err != nil {
			httpx.OkJsonCtx(r.Context(), w, response.FailWithInfo(response.InvalidRequestParamCodeInHandler, "Error while handle call", err.Error()))
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}

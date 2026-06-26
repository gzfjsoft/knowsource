package handler

import (
	"net/http"

	"knowsource/api/internal/logic"
	"knowsource/api/internal/svc"
	"knowsource/api/internal/types"
	"knowsource/common/response"

	"github.com/zeromicro/go-zero/rest/httpx"
)

func SendphonecodeHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		ip := getClientIP(r)

		var req types.SendphonecodeRequest
		if err := httpx.Parse(r, &req); err != nil {
			httpx.OkJsonCtx(r.Context(), w, response.FailWithInfo(response.InvalidRequestParamCodeInHandler, "Error while handle call", err.Error()))
			return
		}

		l := logic.NewSendphonecodeLogic(r.Context(), svcCtx)
		resp, err := l.Sendphonecode(&req, ip)
		if err != nil {
			httpx.OkJsonCtx(r.Context(), w, response.FailWithInfo(response.InvalidRequestParamCodeInHandler, "Error while handle call", err.Error()))
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}

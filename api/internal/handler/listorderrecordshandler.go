package handler

import (
	"net/http"

	"knowsource/api/internal/logic"
	"knowsource/api/internal/svc"
	"knowsource/api/internal/types"
	"knowsource/common/response"

	"github.com/zeromicro/go-zero/rest/httpx"
)

func ListOrderRecordsHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.OrderRecordsListRequest
		if err := httpx.Parse(r, &req); err != nil {
			httpx.OkJsonCtx(r.Context(), w, response.FailWithInfo(response.InvalidRequestParamCodeInHandler, "Error while handle call", err.Error()))
			return
		}

		l := logic.NewListOrderRecordsLogic(r.Context(), svcCtx)
		httpx.OkJsonCtx(r.Context(), w, l.ListOrderRecords(&req))
	}
}

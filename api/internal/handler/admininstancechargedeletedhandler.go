package handler

import (
	"net/http"

	"knowsource/api/internal/logic"
	"knowsource/api/internal/svc"

	"github.com/zeromicro/go-zero/rest/httpx"
)

func AdminInstanceChargeDeletedHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l := logic.NewAdminInstanceChargeDeletedLogic(r.Context(), svcCtx)
		resp, err := l.AdminInstanceChargeDeleted()
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}

package handler

import (
	"net/http"

	"knowsource/api/internal/logic"
	"knowsource/api/internal/svc"

	"github.com/zeromicro/go-zero/rest/httpx"
)

func UserHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l := logic.NewUserLogic(r.Context(), svcCtx)
		httpx.OkJsonCtx(r.Context(), w, l.User())
	}
}

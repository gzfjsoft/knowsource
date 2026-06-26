package knowsource

import (
	"net/http"

	"knowsource/api/internal/logic/knowsource"
	"knowsource/api/internal/svc"
	"knowsource/api/internal/types"

	"github.com/zeromicro/go-zero/rest/httpx"
)

// 员工登录
func KnowsourceOALoginHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.KnowsourceOALoginRequest
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := knowsource.NewKnowsourceOALoginLogic(r.Context(), svcCtx)
		resp, err := l.KnowsourceOALogin(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}

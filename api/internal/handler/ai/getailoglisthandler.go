package ai

import (
	"net/http"

	"knowsource/api/internal/logic/ai"
	"knowsource/api/internal/svc"

	"github.com/zeromicro/go-zero/rest/httpx"
)

// 管理员获取ailog.txt list
func GetAiLogListHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l := ai.NewGetAiLogListLogic(r.Context(), svcCtx)
		resp, err := l.GetAiLogList()
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}

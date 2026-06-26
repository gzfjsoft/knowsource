package knowdata

import (
	"knowsource/api/internal/utils"
	"net/http"

	"knowsource/api/internal/logic/knowdata"
	"knowsource/api/internal/svc"
	"knowsource/api/internal/types"

	"github.com/zeromicro/go-zero/rest/httpx"
)

func DeleteAIConfigHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.PathIdRequest
		if err := utils.ParseRequest(w, r, &req); err != nil {
			return
		}

		l := knowdata.NewDeleteAIConfigLogic(r.Context(), svcCtx)
		resp, err := l.DeleteAIConfig(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}

package knowdata

import (
	"knowsource/api/internal/utils"
	"net/http"

	"knowsource/api/internal/logic/knowdata"
	"knowsource/api/internal/svc"
	"knowsource/api/internal/types"

	"github.com/zeromicro/go-zero/rest/httpx"
)

func UpdateAIConfigHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.KnowdataUpdateAIConfigRequest
		if err := utils.ParseRequest(w, r, &req); err != nil {
			return
		}

		l := knowdata.NewUpdateAIConfigLogic(r.Context(), svcCtx)
		resp, err := l.UpdateAIConfig(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}

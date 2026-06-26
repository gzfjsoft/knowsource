package knowdata

import (
	"knowsource/api/internal/utils"
	"net/http"

	"knowsource/api/internal/logic/knowdata"
	"knowsource/api/internal/svc"
	"knowsource/api/internal/types"

	"github.com/zeromicro/go-zero/rest/httpx"
)

func CreateAIConfigHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.KnowdataCreateAIConfigRequest
		if err := utils.ParseRequest(w, r, &req); err != nil {
			return
		}

		l := knowdata.NewCreateAIConfigLogic(r.Context(), svcCtx)
		resp, err := l.CreateAIConfig(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}

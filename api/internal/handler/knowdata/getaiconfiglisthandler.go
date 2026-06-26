package knowdata

import (
	"knowsource/api/internal/utils"
	"net/http"

	"knowsource/api/internal/logic/knowdata"
	"knowsource/api/internal/svc"
	"knowsource/api/internal/types"

	"github.com/zeromicro/go-zero/rest/httpx"
)

func GetAIConfigListHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.KnowdataAIConfigListRequest
		if err := utils.ParseRequest(w, r, &req); err != nil {
			return
		}

		l := knowdata.NewGetAIConfigListLogic(r.Context(), svcCtx)
		resp, err := l.GetAIConfigList(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}

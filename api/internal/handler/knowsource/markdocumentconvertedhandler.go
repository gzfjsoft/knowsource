package knowsource

import (
	"net/http"

	"knowsource/api/internal/logic/knowsource"
	"knowsource/api/internal/svc"
	"knowsource/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/rest/httpx"
)

// 标注文档已转化
func MarkDocumentConvertedHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logx.Info("MarkDocumentConvertedHandler Request: ", r.RequestURI, " Body: ", r.Body)
		var req types.MarkDocumentConvertedRequest
		if err := httpx.Parse(r, &req); err != nil {
			logx.Error("MarkDocumentConvertedHandler Parse error: ", err)
			logx.Info("MarkDocumentConvertedHandler Request: ", r.RequestURI, " Body: ", r.Body)
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		logx.Info(req)

		l := knowsource.NewMarkDocumentConvertedLogic(r.Context(), svcCtx)
		resp, err := l.MarkDocumentConverted(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}

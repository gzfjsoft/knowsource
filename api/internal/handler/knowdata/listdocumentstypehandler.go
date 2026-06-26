package knowdata

import (
	"net/http"

	"knowsource/api/internal/logic/knowdata"
	"knowsource/api/internal/svc"

	"github.com/zeromicro/go-zero/rest/httpx"
)

// 获取文档类型列表
func ListDocumentsTypeHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l := knowdata.NewListDocumentsTypeLogic(r.Context(), svcCtx)
		resp, err := l.ListDocumentsType()
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}

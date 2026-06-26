package knowdata

import (
	"net/http"

	"knowsource/api/internal/logic/knowdata"
	"knowsource/api/internal/svc"

	"github.com/zeromicro/go-zero/rest/httpx"
)

// 检查所有原始文档文件是否存在于硬盘
func CheckRawDocumentsFileExistsHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l := knowdata.NewCheckRawDocumentsFileExistsLogic(r.Context(), svcCtx)
		resp, err := l.CheckRawDocumentsFileExists()
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}

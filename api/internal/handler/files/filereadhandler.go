package files

import (
	"net/http"
	"path"

	"knowsource/api/internal/svc"
	"knowsource/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/rest/httpx"
)

func FileReadHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		var req types.FileReadRequest
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		filePath := path.Join(svcCtx.Config.FilesRoot, req.File)

		logx.Infof("filePath: %s", filePath)
		http.ServeFile(w, r, filePath)

	}
}

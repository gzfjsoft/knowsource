package mysql

import (
	"net/http"

	mysqllogic "knowsource/api/internal/logic/mysql"
	"knowsource/api/internal/svc"

	"github.com/zeromicro/go-zero/rest/httpx"
)

func MysqlBackupHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l := mysqllogic.NewMysqlBackupLogic(r.Context(), svcCtx)
		resp, err := l.MysqlBackup()
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}
		httpx.OkJsonCtx(r.Context(), w, resp)
	}
}

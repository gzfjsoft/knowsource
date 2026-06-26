package mysql

import (
	"mime"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	mysqllogic "knowsource/api/internal/logic/mysql"
	"knowsource/api/internal/svc"
	"knowsource/api/internal/types"
	"knowsource/api/internal/utils"
	"knowsource/consts"

	"github.com/zeromicro/go-zero/rest/httpx"
)

var mysqlBackupZipName = regexp.MustCompile(`^mysql_backup_\d{14}\.zip$`)

// 下载备份 zip（Auth 中间件已限制 admin 租户 + superadmin；此处再兜底校验）
func MysqlBackupDownloadHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.MysqlBackupDownloadRequest
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		name := strings.TrimSpace(req.File)
		if name == "" || name != filepath.Base(name) || strings.Contains(name, "..") {
			http.Error(w, "非法文件名", http.StatusBadRequest)
			return
		}
		if !mysqlBackupZipName.MatchString(name) {
			http.Error(w, "非法文件名", http.StatusBadRequest)
			return
		}

		clientId, _ := r.Context().Value("clientId").(string)
		if !strings.EqualFold(strings.TrimSpace(clientId), consts.ONLY_ADMIN) || !utils.IsSuperAdminRoleFromContext(r.Context()) {
			http.Error(w, "权限不足，仅 admin 租户的 superadmin 角色可下载", http.StatusForbidden)
			return
		}

		outDir, err := mysqllogic.ResolveBackupOutputDir()
		if err != nil {
			http.Error(w, "备份目录不可用", http.StatusInternalServerError)
			return
		}
		full := filepath.Join(outDir, name)
		rel, err := filepath.Rel(outDir, full)
		if err != nil || strings.HasPrefix(rel, "..") {
			http.Error(w, "非法路径", http.StatusBadRequest)
			return
		}
		st, err := os.Stat(full)
		if err != nil || st.IsDir() {
			http.Error(w, "文件不存在", http.StatusNotFound)
			return
		}

		if ct := mime.TypeByExtension(".zip"); ct != "" {
			w.Header().Set("Content-Type", ct)
		} else {
			w.Header().Set("Content-Type", "application/zip")
		}
		asciiFallback := strings.ReplaceAll(name, `"`, "")
		if strings.TrimSpace(asciiFallback) == "" {
			asciiFallback = "backup.zip"
		}
		escaped := url.PathEscape(name)
		w.Header().Set("Content-Disposition", `attachment; filename="`+asciiFallback+`"; filename*=UTF-8''`+escaped)
		http.ServeFile(w, r, full)
	}
}

package knowdata

import (
	"mime"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"knowsource/api/internal/svc"
	"knowsource/api/internal/types"
	"knowsource/consts"

	"github.com/zeromicro/go-zero/rest/httpx"
)

// 下载原始文档源文件
func DownloadRawDocumentsHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.DownloadRawDocumentsRequest
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		if req.Id <= 0 {
			http.Error(w, "参数错误：id 不能为空", http.StatusBadRequest)
			return
		}

		clientId, _ := r.Context().Value("clientId").(string)
		clientId = strings.TrimSpace(clientId)
		if clientId == "" {
			http.Error(w, "clientId不能为空，请重新登录", http.StatusUnauthorized)
			return
		}

		// 权限检查：这里额外做一次兜底（主要依赖 Auth middleware）
		isAdmin, _ := r.Context().Value("isAdmin").(int64)
		role, _ := r.Context().Value("role").(string)

		doc, err := svcCtx.RawDocumentsModel.FindOneByClientId(r.Context(), clientId, req.Id)
		if err != nil || doc == nil {
			http.Error(w, "文件不存在（id="+strconv.FormatInt(req.Id, 10)+"）", http.StatusNotFound)
			return
		}

		filePath := strings.TrimSpace(doc.FilePath)
		if filePath == "" {
			http.Error(w, "文件路径为空，无法下载："+doc.FileName, http.StatusNotFound)
			return
		}

		st, err := os.Stat(filePath)
		if err != nil || st.IsDir() {
			http.Error(w, "文件不存在，无法下载："+doc.FileName, http.StatusNotFound)
			return
		}

		filename := doc.FileName
		if strings.TrimSpace(filename) == "" {
			filename = filepath.Base(filePath)
		}

		// 权限失败时提示文件名（按你要求）
		if isAdmin != 1 && !strings.Contains(strings.ToLower(role), consts.ONLY_ADMIN) {
			http.Error(w, "没有权限下载文件："+filename, http.StatusForbidden)
			return
		}

		// Content-Type: infer from extension, fallback to octet-stream.
		if ct := mime.TypeByExtension(strings.ToLower(filepath.Ext(filename))); ct != "" {
			w.Header().Set("Content-Type", ct)
		} else {
			w.Header().Set("Content-Type", "application/octet-stream")
		}

		// Content-Disposition: ensure browser downloads with original name.
		asciiFallback := strings.ReplaceAll(filename, `"`, "")
		if strings.TrimSpace(asciiFallback) == "" {
			asciiFallback = "download"
		}
		escaped := url.PathEscape(filename)
		w.Header().Set("Content-Disposition", `attachment; filename="`+asciiFallback+`"; filename*=UTF-8''`+escaped)
		// 不要手动设置 Content-Length，交给 ServeFile 设置；否则带 Range 的请求会返回 206 但 body 只有一段，浏览器会一直等满原 Content-Length
		http.ServeFile(w, r, filePath)
	}
}

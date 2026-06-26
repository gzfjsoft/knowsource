// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package knowdata

import (
	"context"
	"net/http"

	"knowsource/api/internal/logic/knowdata"
	"knowsource/api/internal/svc"
	"knowsource/api/internal/types"

	"github.com/zeromicro/go-zero/rest/httpx"
)

// 重新生成所有已审核文档的概要
func RegenerateAllSummariesHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.RegenerateAllSummariesRequest
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		// 检查是否是流式请求
		if r.Header.Get("Accept") == "text/event-stream" {
			// 对于流式请求，设置响应头
			w.Header().Set("Content-Type", "text/event-stream")
			w.Header().Set("Cache-Control", "no-cache")
			w.Header().Set("Connection", "keep-alive")

			// 创建一个新的 context，包含 http.ResponseWriter
			ctx := context.WithValue(r.Context(), "http.ResponseWriter", w)
			l := knowdata.NewRegenerateAllSummariesLogic(ctx, svcCtx)
			_, err := l.RegenerateAllSummaries(&req)
			if err != nil {
				httpx.ErrorCtx(ctx, w, err)
			}
		} else {
			// 对于普通请求，使用标准处理
			l := knowdata.NewRegenerateAllSummariesLogic(r.Context(), svcCtx)
			resp, err := l.RegenerateAllSummaries(&req)
			if err != nil {
				httpx.ErrorCtx(r.Context(), w, err)
			} else {
				httpx.OkJsonCtx(r.Context(), w, resp)
			}
		}
	}
}

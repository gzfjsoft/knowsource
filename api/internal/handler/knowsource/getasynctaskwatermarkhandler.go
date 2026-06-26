// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package knowsource

import (
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"
	"knowsource/api/internal/logic/knowsource"
	"knowsource/api/internal/svc"
	"knowsource/api/internal/types"
)

// async_task 变更水印（按租户 Redis），前端轮询与本地比较，有变化再拉列表
func GetAsyncTaskWatermarkHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.GetAsyncTaskWatermarkRequest
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := knowsource.NewGetAsyncTaskWatermarkLogic(r.Context(), svcCtx)
		resp, err := l.GetAsyncTaskWatermark(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}

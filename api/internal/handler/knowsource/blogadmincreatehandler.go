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

// 博客管理创建（仅 admin 租户 superadmin）
func BlogAdminCreateHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.BlogAdminCreateRequest
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := knowsource.NewBlogAdminCreateLogic(r.Context(), svcCtx)
		resp, err := l.BlogAdminCreate(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}

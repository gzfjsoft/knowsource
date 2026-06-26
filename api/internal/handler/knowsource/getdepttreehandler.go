// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package knowsource

import (
	"net/http"

	"knowsource/api/internal/logic/knowsource"
	"knowsource/api/internal/svc"
	"knowsource/api/internal/types"

	"github.com/zeromicro/go-zero/rest/httpx"
)

// 获取部门树
func GetDeptTreeHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.KnowsourceDeptTreeRequest
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := knowsource.NewGetDeptTreeLogic(r.Context(), svcCtx)
		resp, err := l.GetDeptTree(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}

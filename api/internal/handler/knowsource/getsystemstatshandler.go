// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package knowsource

import (
	"net/http"

	"knowsource/api/internal/logic/knowsource"
	"knowsource/api/internal/svc"

	"github.com/zeromicro/go-zero/rest/httpx"
)

// 系统统计：返回员工总数、部门总数、知识库总数、AI 会话数
func GetSystemStatsHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l := knowsource.NewGetSystemStatsLogic(r.Context(), svcCtx)
		resp, err := l.GetSystemStats()
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}

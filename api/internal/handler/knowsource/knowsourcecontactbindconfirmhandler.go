// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package knowsource

import (
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"
	"knowsource/api/internal/logic/knowsource"
	"knowsource/api/internal/svc"
	"knowsource/api/internal/types"
)

// 绑定手机/邮箱：提交验证码
func KnowsourceContactBindConfirmHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.KnowsourceContactBindConfirmRequest
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := knowsource.NewKnowsourceContactBindConfirmLogic(r.Context(), svcCtx)
		resp, err := l.KnowsourceContactBindConfirm(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}

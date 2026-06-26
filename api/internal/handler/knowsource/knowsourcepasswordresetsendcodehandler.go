// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package knowsource

import (
	"net/http"

	"knowsource/api/internal/logic/knowsource"
	"knowsource/api/internal/svc"
	"knowsource/api/internal/types"
	"knowsource/api/internal/utils"

	"github.com/zeromicro/go-zero/rest/httpx"
)

// 忘记密码：发送验证码（邮箱或手机）
func KnowsourcePasswordResetSendCodeHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.KnowsourcePasswordResetSendCodeRequest
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := knowsource.NewKnowsourcePasswordResetSendCodeLogic(r.Context(), svcCtx)
		resp, err := l.KnowsourcePasswordResetSendCode(&req, utils.ClientIP(r))
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}

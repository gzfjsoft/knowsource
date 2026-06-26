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

// 绑定手机/邮箱：发送验证码
func KnowsourceContactBindSendHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.KnowsourceContactBindSendRequest
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := knowsource.NewKnowsourceContactBindSendLogic(r.Context(), svcCtx)
		resp, err := l.KnowsourceContactBindSend(&req, utils.ClientIP(r))
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}

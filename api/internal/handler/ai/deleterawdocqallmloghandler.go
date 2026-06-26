// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package ai

import (
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"
	"knowsource/api/internal/logic/ai"
	"knowsource/api/internal/svc"
	"knowsource/api/internal/types"
)

// 管理员删除问答抽取LLM日志
func DeleteRawDocQaLlmLogHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.PathNameRequest
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := ai.NewDeleteRawDocQaLlmLogLogic(r.Context(), svcCtx)
		resp, err := l.DeleteRawDocQaLlmLog(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}

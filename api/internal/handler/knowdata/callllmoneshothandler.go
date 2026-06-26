// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package knowdata

import (
	"net/http"

	"knowsource/api/internal/logic/knowdata"
	"knowsource/api/internal/svc"
	"knowsource/api/internal/types"

	"github.com/zeromicro/go-zero/rest/httpx"
)

// 调用 LLM 进行单轮对话
func CallLLMOneShotHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.CallLLMOneShotRequest
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := knowdata.NewCallLLMOneShotLogic(r.Context(), svcCtx)
		resp, err := l.CallLLMOneShot(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			if svcCtx.Config.Llm.CompletionType == "ollama" {
				w.Header().Set("llmbackend", "ollama")
			}
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}

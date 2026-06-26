// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package ai

import (
	"net/http"

	"knowsource/api/internal/logic/ai"
	"knowsource/api/internal/svc"
	"knowsource/api/internal/types"

	"github.com/zeromicro/go-zero/rest/httpx"
)

// AI对话删除已上传的临时文档（按文件名从缓存移除）
func AIChatRemoveDocumentHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.AIChatRemoveDocumentRequest
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := ai.NewAIChatRemoveDocumentLogic(r.Context(), svcCtx)
		resp, err := l.AIChatRemoveDocument(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}

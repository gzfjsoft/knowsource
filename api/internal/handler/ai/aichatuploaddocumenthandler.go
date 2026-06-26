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

// AI对话上传临时文档（txt/docx/pdf），识别内容并写入对话缓存，发下一条消息时作为参考并清除缓存
func AIChatUploadDocumentHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// 本接口为 multipart/form-data 上传，不解析 JSON
		l := ai.NewAIChatUploadDocumentLogic(r.Context(), svcCtx)
		l.SetRequest(r)
		resp, err := l.AIChatUploadDocument(&types.AIChatUploadDocumentRequest{})
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}

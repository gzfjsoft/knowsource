// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package knowsource

import (
	"net/http"

	"knowsource/api/internal/logic/knowsource"
	"knowsource/api/internal/svc"

	"github.com/zeromicro/go-zero/rest/httpx"
)

// 获取 Embedding 模型列表（从 Rag.EmbeddingsUrl 拉取），仅返回名称含 embedding 的模型
func GetLLMEmbeddingModelsHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l := knowsource.NewGetLLMEmbeddingModelsLogic(r.Context(), svcCtx)
		resp, err := l.GetLLMEmbeddingModels()
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}

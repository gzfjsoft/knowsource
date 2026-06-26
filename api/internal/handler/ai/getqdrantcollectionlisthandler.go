// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package ai

import (
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"
	"knowsource/api/internal/logic/ai"
	"knowsource/api/internal/svc"
)

// 超级管理员获取Qdrant集合列表
func GetQdrantCollectionListHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l := ai.NewGetQdrantCollectionListLogic(r.Context(), svcCtx)
		resp, err := l.GetQdrantCollectionList()
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}

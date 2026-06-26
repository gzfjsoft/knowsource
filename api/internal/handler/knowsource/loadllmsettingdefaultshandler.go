// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package knowsource

import (
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"
	"knowsource/api/internal/logic/knowsource"
	"knowsource/api/internal/svc"
)

// 获取 LLM 设置系统默认值（来自 knowsource.yaml，供界面重置，不写入租户配置）
func LoadLLMSettingDefaultsHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l := knowsource.NewLoadLLMSettingDefaultsLogic(r.Context(), svcCtx)
		resp, err := l.LoadLLMSettingDefaults()
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}

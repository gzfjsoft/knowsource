package bootstrap

import (
	"net/http"

	"github.com/zeromicro/go-zero/rest"
)

// RegisterRoutes 注册无鉴权的初始化与配置接口（需在 RegisterHandlers 之前或之后均可，建议使用独立前缀避免冲突）。
func RegisterRoutes(server *rest.Server, opts Options) {
	server.AddRoutes(
		[]rest.Route{
			{
				Method:  http.MethodGet,
				Path:    "/sys/bootstrap/status",
				Handler: statusHandler(opts),
			},
			{
				Method:  http.MethodGet,
				Path:    "/sys/bootstrap/config",
				Handler: getConfigHandler(opts),
			},
			{
				Method:  http.MethodPost,
				Path:    "/sys/bootstrap/config",
				Handler: saveConfigHandler(opts),
			},
		},
		rest.WithPrefix("/api/v1"),
	)
}

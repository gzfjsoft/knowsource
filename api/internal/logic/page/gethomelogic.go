package page

import (
	"context"

	"knowsource/api/internal/svc"
	"knowsource/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetHomeLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// Get home page
func NewGetHomeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetHomeLogic {
	return &GetHomeLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

var API_FILE string

func (l *GetHomeLogic) GetHome() (resp *types.HtmlResponse, err error) {

	// body := fmt.Sprintf("<html>	<head>		<title>Page Title</title>	</head>	<body>	<pre>%s</pre>	</body>	</html>", API_FILE)

	body := "<html><head><title>api/v1</title></head><body>api/v1</body></html>"
	return &types.HtmlResponse{
		ContentType: "text/html",
		Body:        body,
	}, nil
}

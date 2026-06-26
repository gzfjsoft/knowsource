package knowdata

import (
	"context"

	"knowsource/api/internal/svc"
	"knowsource/api/internal/types"
	"knowsource/api/version"

	"github.com/zeromicro/go-zero/core/logx"
)

type SysVersionLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewSysVersionLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SysVersionLogic {
	return &SysVersionLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *SysVersionLogic) SysVersion() (resp *types.Response, err error) {
	return &types.Response{
		Code:    200,
		Message: "success",
		Info:    version.Version,
	}, nil
}

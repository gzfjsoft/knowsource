package knowdata

import (
	"context"

	"knowsource/api/internal/svc"
	"knowsource/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type SysPingLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewSysPingLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SysPingLogic {
	return &SysPingLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *SysPingLogic) SysPing() (resp *types.Response, err error) {
	resp = &types.Response{
		Code:    200,
		Message: "pong",
	}
	return
}

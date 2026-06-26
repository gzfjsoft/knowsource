package logic

import (
	"context"

	"knowsource/api/internal/svc"
	"knowsource/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type ResendEmailLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewResendEmailLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ResendEmailLogic {
	return &ResendEmailLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ResendEmailLogic) ResendEmail(req *types.ResendEmailRequest) (resp *types.Response, err error) {
	// todo: add your logic here and delete this line

	return
}

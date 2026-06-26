package files

import (
	"context"

	"knowsource/api/internal/svc"
	"knowsource/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type NavigateDirectoryLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewNavigateDirectoryLogic(ctx context.Context, svcCtx *svc.ServiceContext) *NavigateDirectoryLogic {
	return &NavigateDirectoryLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *NavigateDirectoryLogic) NavigateDirectory(req *types.NavigateDirectoryRequest) (resp *types.ListDirectoryResponse, err error) {
	// todo: add your logic here and delete this line

	return
}

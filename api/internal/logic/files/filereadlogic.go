package files

import (
	"context"

	"knowsource/api/internal/svc"
	"knowsource/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type FileReadLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewFileReadLogic(ctx context.Context, svcCtx *svc.ServiceContext) *FileReadLogic {
	return &FileReadLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *FileReadLogic) FileRead(req *types.FileReadRequest) error {
	// todo: add your logic here and delete this line

	return nil
}

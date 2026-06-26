package knowdata

import (
	"context"

	"knowsource/api/internal/svc"
	"knowsource/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type DownloadRawDocumentsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 下载原始文档源文件
func NewDownloadRawDocumentsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DownloadRawDocumentsLogic {
	return &DownloadRawDocumentsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *DownloadRawDocumentsLogic) DownloadRawDocuments(req *types.DownloadRawDocumentsRequest) (resp *types.Response, err error) {
	// todo: add your logic here and delete this line

	return
}

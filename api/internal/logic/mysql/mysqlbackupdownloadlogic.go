// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package mysql

import (
	"context"

	"knowsource/api/internal/svc"
	"knowsource/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type MysqlBackupDownloadLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 下载备份 zip（仅 admin 租户且 superadmin 角色）
func NewMysqlBackupDownloadLogic(ctx context.Context, svcCtx *svc.ServiceContext) *MysqlBackupDownloadLogic {
	return &MysqlBackupDownloadLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *MysqlBackupDownloadLogic) MysqlBackupDownload(req *types.MysqlBackupDownloadRequest) (resp *types.Response, err error) {
	// todo: add your logic here and delete this line

	return
}

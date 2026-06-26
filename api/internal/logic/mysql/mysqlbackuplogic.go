package mysql

import (
	"archive/zip"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"knowsource/api/internal/svc"
	"knowsource/api/internal/types"
	"knowsource/api/internal/utils"
	"knowsource/common/response"
	"knowsource/consts"

	"github.com/zeromicro/go-zero/core/logx"
)

type MysqlBackupLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewMysqlBackupLogic(ctx context.Context, svcCtx *svc.ServiceContext) *MysqlBackupLogic {
	return &MysqlBackupLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *MysqlBackupLogic) MysqlBackup() (resp *types.MysqlBackupResponse, err error) {
	clientId, _ := l.ctx.Value("clientId").(string)
	if !strings.EqualFold(strings.TrimSpace(clientId), consts.ONLY_ADMIN) || !utils.IsSuperAdminRoleFromContext(l.ctx) {
		return &types.MysqlBackupResponse{
			Response: types.Response{
				Code:    response.UnauthorizedCode,
				Message: "权限不足，仅 admin 租户的 superadmin 角色可执行备份",
			},
		}, nil
	}

	dumpBin, host, user, password, database, port, err := resolveDumpConfig(&l.svcCtx.Config)
	if err != nil {
		l.Errorf("备份配置: %v", err)
		return &types.MysqlBackupResponse{
			Response: types.Response{
				Code:    response.ServerErrorCode,
				Message: "备份配置无效",
				Info:    err.Error(),
			},
		}, nil
	}

	outDir, err := ResolveBackupOutputDir()
	if err != nil {
		l.Errorf("输出目录: %v", err)
		return &types.MysqlBackupResponse{
			Response: types.Response{
				Code:    response.ServerErrorCode,
				Message: "无法准备备份目录",
				Info:    err.Error(),
			},
		}, nil
	}

	ts := time.Now().Format("20060102150405")
	base := fmt.Sprintf("mysql_backup_%s", ts)
	sqlName := base + ".sql"
	zipName := base + ".zip"
	sqlPath := filepath.Join(outDir, sqlName)
	zipPath := filepath.Join(outDir, zipName)

	sqlFile, err := os.OpenFile(sqlPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0640)
	if err != nil {
		return &types.MysqlBackupResponse{
			Response: types.Response{
				Code:    response.ServerErrorCode,
				Message: "无法创建 SQL 文件",
				Info:    err.Error(),
			},
		}, nil
	}

	portStr := fmt.Sprintf("%d", port)
	args := []string{
		"-h", host,
		"-P", portStr,
		"-u", user,
		"-p" + password,
		"--single-transaction",
		"--routines",
		"--events",
		database,
	}
	cmd := exec.CommandContext(l.ctx, dumpBin, args...)
	cmd.Stdout = sqlFile
	var stderr strings.Builder
	cmd.Stderr = &stderr
	runErr := cmd.Run()
	_ = sqlFile.Close()
	if runErr != nil {
		_ = os.Remove(sqlPath)
		l.Errorf("mysqldump 失败: %v, stderr: %s", runErr, stderr.String())
		return &types.MysqlBackupResponse{
			Response: types.Response{
				Code:    response.ServerErrorCode,
				Message: "mysqldump 执行失败",
				Info:    stderr.String(),
			},
		}, nil
	}

	if err := zipOneFile(sqlPath, zipPath); err != nil {
		_ = os.Remove(sqlPath)
		_ = os.Remove(zipPath)
		l.Errorf("打包 zip 失败: %v", err)
		return &types.MysqlBackupResponse{
			Response: types.Response{
				Code:    response.ServerErrorCode,
				Message: "压缩备份失败",
				Info:    err.Error(),
			},
		}, nil
	}
	_ = os.Remove(sqlPath)

	return &types.MysqlBackupResponse{
		Response: types.Response{
			Code:    response.SuccessCode,
			Message: "success",
		},
		Data: &types.MysqlBackupData{
			FileName: zipName,
		},
	}, nil
}

func zipOneFile(srcPath, zipPath string) error {
	src, err := os.Open(srcPath)
	if err != nil {
		return err
	}
	defer src.Close()

	zf, err := os.Create(zipPath)
	if err != nil {
		return err
	}
	defer zf.Close()

	zw := zip.NewWriter(zf)
	w, err := zw.Create(filepath.Base(srcPath))
	if err != nil {
		_ = zw.Close()
		return err
	}
	if _, err := io.Copy(w, src); err != nil {
		_ = zw.Close()
		return err
	}
	return zw.Close()
}

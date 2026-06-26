package knowsource

import (
	"bytes"
	"context"
	"os/exec"

	"knowsource/api/internal/svc"
	"knowsource/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type SyncMysqlLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// sync mysql
func NewSyncMysqlLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SyncMysqlLogic {
	return &SyncMysqlLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *SyncMysqlLogic) SyncMysql() (resp *types.Response, err error) {
	// 执行 shell 命令
	cmd := exec.Command("/root/code/src.go/knowsource/mssql-sync/hr-mssql")

	// 捕获标准输出和标准错误
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	// 执行命令
	err = cmd.Run()

	// 合并输出
	var output string
	if stdout.Len() > 0 {
		output = stdout.String()
	}
	if stderr.Len() > 0 {
		if output != "" {
			output += "\n" + stderr.String()
		} else {
			output = stderr.String()
		}
	}

	// 如果命令执行失败，记录错误
	if err != nil {
		l.Errorf("Failed to execute hr-mssql: %v, output: %s", err, output)
		return &types.Response{
			Code:    500,
			Message: output,
			Info:    err.Error(),
		}, nil
	}

	// 成功执行
	l.Infof("hr-mssql executed successfully, output length: %d", len(output))
	return &types.Response{
		Code:    200,
		Message: output,
	}, nil
}

package executor

import (
	"context"
	"fmt"
	"knowsource/api/internal/svc"

	"knowsource/model"

	"github.com/zeromicro/go-zero/core/logx"
)

// 示例任务执行器实现
type MakeOfflinePackageExecutor struct {
	svcCtx *svc.ServiceContext
}

func NewMakeOfflinePackageExecutor(svcCtx *svc.ServiceContext) *MakeOfflinePackageExecutor {
	return &MakeOfflinePackageExecutor{
		svcCtx: svcCtx,
	}
}

func (e *MakeOfflinePackageExecutor) Execute(ctx context.Context, task *model.AsyncTask) (string, error) {
	logx.Infof("Executing example task: %s, Description: %s, Status: %s", task.TaskType, task.TaskDesc, task.Status)
	if task.TaskType != "make_offline_package" {
		return "任务类型不匹配", fmt.Errorf("unsupported task type: %s", task.TaskType)
	}

	return "执行成功", nil
}

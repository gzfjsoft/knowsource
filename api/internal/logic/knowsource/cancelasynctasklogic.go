// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package knowsource

import (
	"context"
	"strings"

	"knowsource/api/internal/svc"
	"knowsource/api/internal/types"
	"knowsource/common/asynctasksignal"
	"knowsource/common/constants"
	"knowsource/common/response"
	"knowsource/model"

	"github.com/zeromicro/go-zero/core/logx"
)

type CancelAsyncTaskLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// cancel async task for current tenant
func NewCancelAsyncTaskLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CancelAsyncTaskLogic {
	return &CancelAsyncTaskLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CancelAsyncTaskLogic) CancelAsyncTask(req *types.CancelAsyncTaskRequest) (resp *types.Response, err error) {
	clientId, _ := l.ctx.Value("clientId").(string)
	clientId = strings.TrimSpace(clientId)
	if clientId == "" {
		return &types.Response{
			Code:    response.UnauthorizedCode,
			Message: "clientId不能为空，请重新登录",
		}, nil
	}
	if req.Id <= 0 {
		return &types.Response{
			Code:    response.ParameterErrorCode,
			Message: "id不能为空",
		}, nil
	}

	taskModel := model.NewAsyncTaskModel(l.svcCtx.Mysql)
	task, findErr := taskModel.FindOne(l.ctx, req.Id)
	if findErr != nil || task == nil || strings.TrimSpace(task.ClientId) != clientId {
		return &types.Response{
			Code:    response.ParameterErrorCode,
			Message: "任务不存在",
		}, nil
	}

	status := strings.TrimSpace(task.Status)
	if status != constants.AsyncTaskStatusInit && status != constants.AsyncTaskStatusRunning {
		return &types.Response{
			Code:    response.SuccessCode,
			Message: "任务已结束，无需取消",
		}, nil
	}

	ok, cancelErr := taskModel.CancelActiveById(l.ctx, clientId, req.Id, "用户手工取消任务")
	if cancelErr != nil {
		return &types.Response{
			Code:    response.ServerErrorCode,
			Message: "取消任务失败",
			Info:    cancelErr.Error(),
		}, nil
	}
	if ok {
		RevertAfterAsyncTaskCanceled(l.ctx, l.svcCtx, task)
		asynctasksignal.BumpClientWatermark(l.ctx, l.svcCtx.RedisClient, clientId)
		return &types.Response{
			Code:    response.SuccessCode,
			Message: "已提交取消请求",
		}, nil
	}

	return &types.Response{
		Code:    response.SuccessCode,
		Message: "任务状态已变化，无需取消",
	}, nil
}

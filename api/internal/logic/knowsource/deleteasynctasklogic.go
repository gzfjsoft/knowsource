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

type DeleteAsyncTaskLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// delete async task for current tenant
func NewDeleteAsyncTaskLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteAsyncTaskLogic {
	return &DeleteAsyncTaskLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *DeleteAsyncTaskLogic) DeleteAsyncTask(req *types.DeleteAsyncTaskRequest) (resp *types.Response, err error) {
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
	if status == constants.AsyncTaskStatusInit || status == constants.AsyncTaskStatusRunning {
		return &types.Response{
			Code:    response.ParameterErrorCode,
			Message: "任务正在执行，请先停止后再删除",
		}, nil
	}

	if delErr := taskModel.Delete(l.ctx, req.Id); delErr != nil {
		return &types.Response{
			Code:    response.ServerErrorCode,
			Message: "删除任务失败",
			Info:    delErr.Error(),
		}, nil
	}
	asynctasksignal.BumpClientWatermark(l.ctx, l.svcCtx.RedisClient, clientId)
	return &types.Response{
		Code:    response.SuccessCode,
		Message: "删除成功",
	}, nil
}

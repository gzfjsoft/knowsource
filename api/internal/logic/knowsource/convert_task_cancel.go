package knowsource

import (
	"context"
	"strings"

	"knowsource/api/internal/svc"
	"knowsource/common/asynctask"
	"knowsource/common/constants"
	"knowsource/model"
)

// checkConvertTaskCanceled 识别过程中检测异步任务是否已被取消；若已取消则恢复文档状态
func (l *ConvertDocumentToZIPLogic) checkConvertTaskCanceled(asyncTaskId int64, docId int64) error {
	return checkConvertTaskCanceled(l.ctx, l.svcCtx, asyncTaskId, docId)
}

func (l *ConvertDocumentToMDLogic) checkConvertTaskCanceled(asyncTaskId int64, docId int64) error {
	return checkConvertTaskCanceled(l.ctx, l.svcCtx, asyncTaskId, docId)
}

func checkConvertTaskCanceled(ctx context.Context, svcCtx *svc.ServiceContext, asyncTaskId int64, docId int64) error {
	if asyncTaskId <= 0 {
		return nil
	}
	taskModel := model.NewAsyncTaskModel(svcCtx.Mysql)
	cur, err := taskModel.FindOne(ctx, asyncTaskId)
	if err != nil || cur == nil {
		return nil
	}
	if cur.Status != constants.AsyncTaskStatusCanceled {
		return nil
	}
	clientId, _ := ctx.Value("clientId").(string)
	id := docId
	if id <= 0 {
		id = cur.SourceId
	}
	_, _, _ = RevertRawDocumentAfterCancelExtract(ctx, svcCtx, strings.TrimSpace(clientId), id)
	return asynctask.ErrTaskCancelled
}

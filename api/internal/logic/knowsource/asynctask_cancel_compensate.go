package knowsource

import (
	"context"
	"strings"

	"knowsource/api/internal/svc"
	"knowsource/common/constants"
	"knowsource/model"
)

// RevertAfterAsyncTaskCanceled 根据任务类型做业务补偿，避免文档状态停留在进行中。
func RevertAfterAsyncTaskCanceled(ctx context.Context, svcCtx *svc.ServiceContext, task *model.AsyncTask) {
	if svcCtx == nil || task == nil {
		return
	}
	clientId := strings.TrimSpace(task.ClientId)
	switch strings.TrimSpace(task.TaskType) {
	case constants.AsyncTaskTypeRawDocumentsConvertZIP, constants.AsyncTaskTypeRawDocumentsConvertMD:
		_, _, _ = RevertRawDocumentAfterCancelExtract(ctx, svcCtx, clientId, task.SourceId)
	case constants.AsyncTaskTypeRawDocumentsAuditIn:
		_, _ = RevertRawDocumentAfterCancelAudit(ctx, svcCtx, clientId, task.SourceId)
	default:
		// 其他任务类型暂无补偿逻辑
	}
}

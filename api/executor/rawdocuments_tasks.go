package executor

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	knowdataLogic "knowsource/api/internal/logic/knowdata"
	knowsourceLogic "knowsource/api/internal/logic/knowsource"
	"knowsource/api/internal/svc"
	"knowsource/api/internal/types"
	"knowsource/common/asynctask"
	"knowsource/common/constants"
	"knowsource/model"
)

type RawDocumentsConvertZIPExecutor struct {
	svcCtx    *svc.ServiceContext
	taskModel model.AsyncTaskModel
}

func NewRawDocumentsConvertZIPExecutor(svcCtx *svc.ServiceContext) *RawDocumentsConvertZIPExecutor {
	return &RawDocumentsConvertZIPExecutor{
		svcCtx:    svcCtx,
		taskModel: model.NewAsyncTaskModel(svcCtx.Mysql),
	}
}

func (e *RawDocumentsConvertZIPExecutor) Execute(ctx context.Context, task *model.AsyncTask) (string, error) {
	if task == nil {
		return "task is nil", fmt.Errorf("task is nil")
	}
	clientId := strings.TrimSpace(task.ClientId)
	if clientId != "" {
		ctx = context.WithValue(ctx, "clientId", clientId)
	}

	if cur, _ := e.taskModel.FindOne(ctx, task.Id); cur != nil && cur.Status == constants.AsyncTaskStatusCanceled {
		_, _, _ = knowsourceLogic.RevertRawDocumentAfterCancelExtract(ctx, e.svcCtx, clientId, task.SourceId)
		return "任务已取消", asynctask.ErrTaskCancelled
	}

	_ = knowsourceLogic.MarkRawDocumentExtracting(ctx, e.svcCtx, clientId, task.SourceId)

	l := knowsourceLogic.NewConvertDocumentToZIPLogic(ctx, e.svcCtx)
	resp, err := l.ConvertDocumentToZIPSync(&types.ConvertDocumentToMDRequest{Id: task.SourceId}, task.Id)
	if syncErr := knowsourceLogic.ConvertZIPSyncResult(resp, err); syncErr != nil {
		if errors.Is(syncErr, asynctask.ErrTaskCancelled) {
			return syncErr.Error(), syncErr
		}
		_ = knowsourceLogic.MarkRawDocumentExtractFailed(ctx, e.svcCtx, clientId, task.SourceId, syncErr.Error())
		return syncErr.Error(), syncErr
	}
	_ = knowsourceLogic.MarkRawDocumentExtractSuccess(ctx, e.svcCtx, clientId, task.SourceId)
	return "识别文字完成", nil
}

type RawDocumentsConvertMDExecutor struct {
	svcCtx    *svc.ServiceContext
	taskModel model.AsyncTaskModel
}

func NewRawDocumentsConvertMDExecutor(svcCtx *svc.ServiceContext) *RawDocumentsConvertMDExecutor {
	return &RawDocumentsConvertMDExecutor{
		svcCtx:    svcCtx,
		taskModel: model.NewAsyncTaskModel(svcCtx.Mysql),
	}
}

func (e *RawDocumentsConvertMDExecutor) Execute(ctx context.Context, task *model.AsyncTask) (string, error) {
	if task == nil {
		return "task is nil", fmt.Errorf("task is nil")
	}
	clientId := strings.TrimSpace(task.ClientId)
	if clientId != "" {
		ctx = context.WithValue(ctx, "clientId", clientId)
	}

	if cur, _ := e.taskModel.FindOne(ctx, task.Id); cur != nil && cur.Status == constants.AsyncTaskStatusCanceled {
		_, _, _ = knowsourceLogic.RevertRawDocumentAfterCancelExtract(ctx, e.svcCtx, clientId, task.SourceId)
		return "任务已取消", asynctask.ErrTaskCancelled
	}

	_ = knowsourceLogic.MarkRawDocumentExtracting(ctx, e.svcCtx, clientId, task.SourceId)

	l := knowsourceLogic.NewConvertDocumentToMDLogic(ctx, e.svcCtx)
	resp, err := l.ConvertDocumentToMDSync(&types.ConvertDocumentToMDRequest{Id: task.SourceId}, task.Id)
	if syncErr := knowsourceLogic.ConvertMDSyncResult(resp, err); syncErr != nil {
		if errors.Is(syncErr, asynctask.ErrTaskCancelled) {
			return syncErr.Error(), syncErr
		}
		_ = knowsourceLogic.MarkRawDocumentExtractFailed(ctx, e.svcCtx, clientId, task.SourceId, syncErr.Error())
		return syncErr.Error(), syncErr
	}
	_ = knowsourceLogic.MarkRawDocumentExtractSuccess(ctx, e.svcCtx, clientId, task.SourceId)
	return "识别文字完成", nil
}

type RawDocumentsAuditInExecutor struct {
	svcCtx    *svc.ServiceContext
	taskModel model.AsyncTaskModel
}

func NewRawDocumentsAuditInExecutor(svcCtx *svc.ServiceContext) *RawDocumentsAuditInExecutor {
	return &RawDocumentsAuditInExecutor{
		svcCtx:    svcCtx,
		taskModel: model.NewAsyncTaskModel(svcCtx.Mysql),
	}
}

func (e *RawDocumentsAuditInExecutor) Execute(ctx context.Context, task *model.AsyncTask) (string, error) {
	if task == nil {
		return "task is nil", fmt.Errorf("task is nil")
	}
	clientId := strings.TrimSpace(task.ClientId)
	if clientId != "" {
		ctx = context.WithValue(ctx, "clientId", clientId)
	}

	if cur, _ := e.taskModel.FindOne(ctx, task.Id); cur != nil && cur.Status == constants.AsyncTaskStatusCanceled {
		_, _ = knowsourceLogic.RevertRawDocumentAfterCancelAudit(ctx, e.svcCtx, clientId, task.SourceId)
		return "任务已取消", asynctask.ErrTaskCancelled
	}

	doc, err := e.svcCtx.RawDocumentsModel.FindOneByClientId(ctx, clientId, task.SourceId)
	if err != nil {
		return "文档不存在", err
	}
	if cur, _ := e.taskModel.FindOne(ctx, task.Id); cur != nil && cur.Status == constants.AsyncTaskStatusCanceled {
		_, _ = knowsourceLogic.RevertRawDocumentAfterCancelAudit(ctx, e.svcCtx, clientId, task.SourceId)
		return "任务已取消", asynctask.ErrTaskCancelled
	}

	_, _ = knowsourceLogic.UpdateRawDocumentStatus(ctx, e.svcCtx, clientId, doc.Id, constants.RawDocumentsStatusInserting, "")

	indexLogic := knowdataLogic.NewIndexRawDocumentToQdrantLogic(ctx, e.svcCtx)
	indexResp, indexErr := indexLogic.IndexRawDocumentToQdrant(&types.IndexRawDocumentToQdrantRequest{Id: doc.Id})
	if indexErr != nil || indexResp == nil || indexResp.Code != 200 {
		reason := "入库失败"
		if indexErr != nil {
			reason = indexErr.Error()
		} else if indexResp != nil {
			reason = strings.TrimSpace(indexResp.Message + " " + indexResp.Info)
		}
		_ = knowsourceLogic.MarkRawDocumentAuditFailed(ctx, e.svcCtx, clientId, doc.Id, reason)
		doc.IsAudit = 0
		doc.AuditUser = ""
		doc.AuditedAt = sql.NullTime{Valid: false}
		doc.UpdatedAt = time.Now()
		_ = e.svcCtx.RawDocumentsModel.Update(ctx, doc)
		if indexErr != nil {
			return indexErr.Error(), indexErr
		}
		return reason, fmt.Errorf("index failed")
	}

	qaCount, qaErr := knowdataLogic.BuildAndStoreRawDocumentQAPairs(ctx, e.svcCtx, doc)
	qaResult := ""
	switch {
	case qaErr != nil:
		if strings.Contains(qaErr.Error(), "未抽取到有效问答") {
			qaResult = "问答抽取完成: 未抽取到有效问答(按正常入库处理)"
			break
		}
		// 问答抽取是增强能力，不影响主流程入库结果；异常写入任务执行结果供前端查看
		qaResult = fmt.Sprintf("问答抽取异常(不影响入库): %v", qaErr)
	case qaCount <= 0:
		qaResult = "问答抽取完成: 未抽取到有效问答(按正常入库处理)"
	default:
		qaResult = fmt.Sprintf("问答抽取完成: %d 条", qaCount)
	}

	if cur, _ := e.taskModel.FindOne(ctx, task.Id); cur != nil && cur.Status == constants.AsyncTaskStatusCanceled {
		_, _ = knowsourceLogic.RevertRawDocumentAfterCancelAudit(ctx, e.svcCtx, clientId, task.SourceId)
		return "任务已取消", asynctask.ErrTaskCancelled
	}

	doc.IsAudit = 1
	doc.AuditUser = strings.TrimSpace(task.SourceKey)
	doc.AuditedAt = sql.NullTime{Time: time.Now(), Valid: true}
	doc.Status = constants.RawDocumentsStatusInserted
	doc.StatusMsg = ""
	doc.UpdatedAt = time.Now()
	if upErr := e.svcCtx.RawDocumentsModel.Update(ctx, doc); upErr != nil {
		return upErr.Error(), upErr
	}

	return strings.TrimSpace(fmt.Sprintf("%s；%s", indexResp.Info, qaResult)), nil
}

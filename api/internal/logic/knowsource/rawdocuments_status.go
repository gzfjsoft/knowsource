package knowsource

import (
	"context"
	"fmt"
	"strings"

	"knowsource/api/internal/svc"
	"knowsource/api/internal/types"
	"knowsource/common/asynctasksignal"
	"knowsource/common/constants"
	"knowsource/model"

	"github.com/zeromicro/go-zero/core/logx"
)

const rawDocumentStatusMsgMaxLen = 500

// TruncateRawDocumentStatusMsg 截断状态说明，避免超出库字段长度
func TruncateRawDocumentStatusMsg(msg string) string {
	msg = strings.TrimSpace(msg)
	if msg == "" {
		return ""
	}
	r := []rune(msg)
	if len(r) <= rawDocumentStatusMsgMaxLen {
		return msg
	}
	return string(r[:rawDocumentStatusMsgMaxLen])
}

// UpdateRawDocumentStatus 更新文档 status 与 status_msg；status_msg 列不存在时回退为仅更新 status
func UpdateRawDocumentStatus(ctx context.Context, svcCtx *svc.ServiceContext, clientId string, docId int64, status, statusMsg string) (int64, error) {
	if svcCtx == nil || docId <= 0 {
		return 0, nil
	}
	clientId = strings.TrimSpace(clientId)
	status = strings.TrimSpace(status)
	statusMsg = TruncateRawDocumentStatusMsg(statusMsg)

	rows, err := svcCtx.RawDocumentsModel.UpdateStatusAndMsg(ctx, clientId, docId, status, statusMsg)
	if err == nil && rows > 0 {
		return rows, nil
	}
	if err != nil {
		logx.WithContext(ctx).Infof("[rawdoc-status] UpdateStatusAndMsg failed, fallback UpdateStatusOnly: docId=%d err=%v", docId, err)
	}
	rowsOnly, errOnly := svcCtx.RawDocumentsModel.UpdateStatusOnly(ctx, clientId, docId, status)
	if errOnly != nil {
		if err != nil {
			return 0, fmt.Errorf("UpdateStatusAndMsg: %v; UpdateStatusOnly: %w", err, errOnly)
		}
		return 0, errOnly
	}
	if rowsOnly == 0 && rows == 0 {
		// client_id 不匹配时再按 id 更新（与 gen Update 一致，仅 status）
		rowsOnly, errOnly = svcCtx.RawDocumentsModel.UpdateStatusOnly(ctx, "", docId, status)
		if errOnly != nil {
			return 0, errOnly
		}
	}
	if statusMsg == "" && rowsOnly > 0 {
		if clearRows, clearErr := svcCtx.RawDocumentsModel.ClearStatusMsg(ctx, clientId, docId); clearErr == nil && clearRows == 0 && clientId != "" {
			_, _ = svcCtx.RawDocumentsModel.ClearStatusMsg(ctx, "", docId)
		}
	}
	return rowsOnly, nil
}

// MarkRawDocumentExtractQueued 上传后/提交识别：等待 worker，展示「已上传，正在提取文字...」
func MarkRawDocumentExtractQueued(ctx context.Context, svcCtx *svc.ServiceContext, clientId string, docId int64) error {
	_, err := UpdateRawDocumentStatus(ctx, svcCtx, clientId, docId, constants.RawDocumentsStatusUploadedExtracting, "")
	return err
}

// MarkRawDocumentExtracting worker 开始执行识别
func MarkRawDocumentExtracting(ctx context.Context, svcCtx *svc.ServiceContext, clientId string, docId int64) error {
	_, err := UpdateRawDocumentStatus(ctx, svcCtx, clientId, docId, constants.RawDocumentsStatusExtracting, "")
	return err
}

// MarkRawDocumentExtractSuccess 识别完成
func MarkRawDocumentExtractSuccess(ctx context.Context, svcCtx *svc.ServiceContext, clientId string, docId int64) error {
	_, err := UpdateRawDocumentStatus(ctx, svcCtx, clientId, docId, constants.RawDocumentsStatusExtractedNotInDB, "")
	return err
}

// MarkRawDocumentExtractFailed 识别失败，可重试
func MarkRawDocumentExtractFailed(ctx context.Context, svcCtx *svc.ServiceContext, clientId string, docId int64, reason string) error {
	if reason == "" {
		reason = "识别失败"
	}
	_, err := UpdateRawDocumentStatus(ctx, svcCtx, clientId, docId, constants.RawDocumentsStatusExtractFailed, reason)
	return err
}

// MarkRawDocumentAuditFailed 审核入库失败
func MarkRawDocumentAuditFailed(ctx context.Context, svcCtx *svc.ServiceContext, clientId string, docId int64, reason string) error {
	if reason == "" {
		reason = "入库失败"
	}
	_, err := UpdateRawDocumentStatus(ctx, svcCtx, clientId, docId, constants.RawDocumentsStatusInsertFailed, reason)
	return err
}

// CancelActiveConvertTasks 取消识别相关 async_task（先按 client_id，再兜底不限 client_id）
func CancelActiveConvertTasks(ctx context.Context, svcCtx *svc.ServiceContext, clientId string, docId int64, result string) (cancelledZip bool, cancelledMD bool, cancelledAny bool, err error) {
	taskModel := model.NewAsyncTaskModel(svcCtx.Mysql)
	for _, taskType := range []string{
		constants.AsyncTaskTypeRawDocumentsConvertZIP,
		constants.AsyncTaskTypeRawDocumentsConvertMD,
	} {
		var ok bool
		if strings.TrimSpace(clientId) != "" {
			ok, err = taskModel.CancelActiveByTaskTypeAndSourceId(ctx, clientId, taskType, docId, result)
			if err != nil {
				return false, false, false, err
			}
		}
		if !ok {
			ok, err = taskModel.CancelActiveByTaskTypeAndSourceIdAnyClient(ctx, taskType, docId, result)
			if err != nil {
				return false, false, false, err
			}
		}
		if ok {
			cancelledAny = true
			switch taskType {
			case constants.AsyncTaskTypeRawDocumentsConvertZIP:
				cancelledZip = true
			case constants.AsyncTaskTypeRawDocumentsConvertMD:
				cancelledMD = true
			}
		}
	}
	return cancelledZip, cancelledMD, cancelledAny, nil
}

// EnqueueRawDocumentConvertZIP 创建识别异步任务并唤醒 worker（上传自动识别 / 手动「识别文字」共用）
func EnqueueRawDocumentConvertZIP(ctx context.Context, svcCtx *svc.ServiceContext, clientId string, docId int64, fileName string) error {
	clientId = strings.TrimSpace(clientId)
	if svcCtx == nil || docId <= 0 {
		return fmt.Errorf("invalid enqueue params")
	}
	taskModel := model.NewAsyncTaskModel(svcCtx.Mysql)
	active, err := taskModel.FindActiveByTaskTypeAndSourceId(ctx, clientId, constants.AsyncTaskTypeRawDocumentsConvertZIP, docId)
	if err != nil {
		return err
	}
	if active != nil {
		return nil
	}
	if err := MarkRawDocumentExtractQueued(ctx, svcCtx, clientId, docId); err != nil {
		return err
	}
	_, err = taskModel.CreateWithClientId(
		ctx,
		clientId,
		constants.AsyncTaskTypeRawDocumentsConvertZIP,
		fmt.Sprintf("识别文字:%s", fileName),
		docId,
		fileName,
	)
	if err != nil {
		return err
	}
	_ = asynctasksignal.NotifyPending(ctx, svcCtx.RedisClient, clientId)
	return nil
}

// ConvertZIPSyncResult 将 ConvertDocumentToZIPSync 的响应转为 executor 可用的 error
func ConvertZIPSyncResult(resp *types.ConvertDocumentToZIPResponse, err error) error {
	if err != nil {
		return err
	}
	if resp == nil {
		return fmt.Errorf("识别返回为空")
	}
	if resp.Code == 200 {
		return nil
	}
	msg := strings.TrimSpace(resp.Message)
	if info := strings.TrimSpace(resp.Info); info != "" {
		if msg != "" {
			msg += ": " + info
		} else {
			msg = info
		}
	}
	if msg == "" {
		msg = "识别失败"
	}
	return fmt.Errorf("%s", msg)
}

// ConvertMDSyncResult 将 ConvertDocumentToMDSync 的响应转为 executor 可用的 error
func ConvertMDSyncResult(resp *types.ConvertDocumentToMDResponse, err error) error {
	if err != nil {
		return err
	}
	if resp == nil {
		return fmt.Errorf("识别返回为空")
	}
	if resp.Code == 200 {
		return nil
	}
	msg := strings.TrimSpace(resp.Message)
	if info := strings.TrimSpace(resp.Info); info != "" {
		if msg != "" {
			msg += ": " + info
		} else {
			msg = info
		}
	}
	if msg == "" {
		msg = "识别失败"
	}
	return fmt.Errorf("%s", msg)
}

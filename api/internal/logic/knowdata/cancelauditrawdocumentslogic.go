// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package knowdata

import (
	"context"
	"strings"

	knowsourceLogic "knowsource/api/internal/logic/knowsource"
	"knowsource/api/internal/svc"
	"knowsource/api/internal/types"
	"knowsource/common/asynctasksignal"
	"knowsource/common/constants"
	"knowsource/common/response"
	"knowsource/model"

	"github.com/zeromicro/go-zero/core/logx"
)

type CancelAuditRawDocumentsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 取消正在进行的审核入库任务
func NewCancelAuditRawDocumentsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CancelAuditRawDocumentsLogic {
	return &CancelAuditRawDocumentsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CancelAuditRawDocumentsLogic) CancelAuditRawDocuments(req *types.CancelAuditRawDocumentsRequest) (resp *types.Response, err error) {
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

	doc, docErr := l.svcCtx.RawDocumentsModel.FindOneByClientId(l.ctx, clientId, req.Id)
	if docErr != nil {
		return &types.Response{
			Code:    response.ServerErrorCode,
			Message: "查询文档失败",
			Info:    docErr.Error(),
		}, nil
	}
	if doc == nil {
		return &types.Response{
			Code:    response.ParameterErrorCode,
			Message: "文档不存在",
		}, nil
	}

	busyInserting := constants.IsRawDocumentInsertingStatus(doc.Status) || strings.TrimSpace(doc.Status) == constants.RawDocumentsStatusInsertFailed
	taskModel := model.NewAsyncTaskModel(l.svcCtx.Mysql)
	activeBefore, aErr := taskModel.FindActiveByTaskTypeAndSourceId(l.ctx, clientId, constants.AsyncTaskTypeRawDocumentsAuditIn, req.Id)
	if aErr != nil {
		return &types.Response{
			Code:    response.ServerErrorCode,
			Message: "查询审核任务失败",
			Info:    aErr.Error(),
		}, nil
	}

	if !busyInserting && activeBefore == nil {
		return &types.Response{
			Code:    response.SuccessCode,
			Message: "当前文档未在入库中",
		}, nil
	}

	cancelled, cErr := taskModel.CancelActiveByTaskTypeAndSourceId(l.ctx, clientId, constants.AsyncTaskTypeRawDocumentsAuditIn, req.Id, "用户取消审核入库任务")
	if cErr != nil {
		return &types.Response{
			Code:    response.ServerErrorCode,
			Message: "取消失败",
			Info:    cErr.Error(),
		}, nil
	}

	rows, revertErr := knowsourceLogic.RevertRawDocumentAfterCancelAudit(l.ctx, l.svcCtx, clientId, req.Id)
	if revertErr != nil {
		l.Errorf("恢复文档状态失败: %v, docId=%d", revertErr, req.Id)
		return &types.Response{
			Code:    response.ServerErrorCode,
			Message: "恢复文档状态失败",
			Info:    revertErr.Error(),
		}, nil
	}
	if rows == 0 {
		l.Errorf("[cancel-audit] revert rows=0 docId=%d", req.Id)
		return &types.Response{
			Code:    response.ServerErrorCode,
			Message: "恢复文档状态失败：未更新任何记录",
		}, nil
	}

	asynctasksignal.BumpClientWatermark(l.ctx, l.svcCtx.RedisClient, clientId)

	revertLabel := constants.RawDocumentsStatusExtractedNotInDB
	msg := "已恢复文档状态为「" + revertLabel + "」"
	if cancelled {
		msg = "已提交取消请求，" + msg
	} else {
		msg = "未发现进行中的审核任务（可能服务已重启），" + msg
	}

	return &types.Response{
		Code:    response.SuccessCode,
		Message: msg,
	}, nil
}

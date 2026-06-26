// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package knowsource

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"knowsource/api/internal/svc"
	"knowsource/api/internal/types"
	"knowsource/common/asynctasksignal"
	"knowsource/common/constants"
	"knowsource/common/response"
	"knowsource/model"

	"github.com/zeromicro/go-zero/core/logx"
)

type CancelConvertDocumentLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// cancel convert task
func NewCancelConvertDocumentLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CancelConvertDocumentLogic {
	return &CancelConvertDocumentLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

type cancelConvertDebug struct {
	DocId              int64  `json:"docId"`
	ClientId           string `json:"clientId"`
	DocClientId        string `json:"docClientId"`
	StatusDB           string `json:"statusDb"`
	StatusDisplay      string `json:"statusDisplay"`
	IsToMd             int64  `json:"isToMd"`
	BusyExtracting     bool   `json:"busyExtracting"`
	ActiveTaskBefore   bool   `json:"activeTaskBefore"`
	CancelledZip       bool   `json:"cancelledZip"`
	CancelledMD        bool   `json:"cancelledMd"`
	RevertTargetStatus string `json:"revertTargetStatus"`
	RevertRowsAffected int64  `json:"revertRowsAffected"`
	StatusAfter        string `json:"statusAfter"`
}

func (l *CancelConvertDocumentLogic) hasActiveConvertTask(clientId string, docId int64) (bool, error) {
	taskModel := model.NewAsyncTaskModel(l.svcCtx.Mysql)
	for _, taskType := range []string{
		constants.AsyncTaskTypeRawDocumentsConvertZIP,
		constants.AsyncTaskTypeRawDocumentsConvertMD,
	} {
		if strings.TrimSpace(clientId) != "" {
			active, err := taskModel.FindActiveByTaskTypeAndSourceId(l.ctx, clientId, taskType, docId)
			if err != nil {
				return false, err
			}
			if active != nil {
				return true, nil
			}
		}
		active, err := taskModel.FindByTaskTypeAndSourceId(l.ctx, taskType, docId)
		if err != nil {
			return false, err
		}
		if active != nil && (active.Status == constants.AsyncTaskStatusInit || active.Status == constants.AsyncTaskStatusRunning) {
			return true, nil
		}
	}
	return false, nil
}

func (l *CancelConvertDocumentLogic) CancelConvertDocument(req *types.CancelConvertDocumentRequest) (resp *types.Response, err error) {
	clientId, _ := l.ctx.Value("clientId").(string)
	clientId = strings.TrimSpace(clientId)
	dbg := cancelConvertDebug{DocId: req.Id, ClientId: clientId}

	l.Infof("[cancel-convert] start docId=%d clientId=%q", req.Id, clientId)

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
		l.Errorf("[cancel-convert] FindOneByClientId failed docId=%d clientId=%q err=%v", req.Id, clientId, docErr)
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

	dbg.DocClientId = strings.TrimSpace(doc.ClientId)
	dbg.StatusDB = strings.TrimSpace(doc.Status)
	dbg.StatusDisplay = constants.ResolveRawDocumentListStatus(doc.Status, doc.IsAudit, doc.IsToMd)
	dbg.IsToMd = doc.IsToMd
	dbg.BusyExtracting = constants.IsRawDocumentExtractingBusy(doc.Status, doc.IsToMd)

	activeBefore, aErr := l.hasActiveConvertTask(clientId, req.Id)
	dbg.ActiveTaskBefore = activeBefore
	if aErr != nil {
		l.Errorf("[cancel-convert] hasActiveConvertTask err=%v", aErr)
		return &types.Response{
			Code:    response.ServerErrorCode,
			Message: "查询识别任务失败",
			Info:    aErr.Error(),
		}, nil
	}

	l.Infof("[cancel-convert] docId=%d statusDb=%q statusDisplay=%q isToMd=%d busy=%v activeTask=%v",
		req.Id, dbg.StatusDB, dbg.StatusDisplay, doc.IsToMd, dbg.BusyExtracting, activeBefore)

	// 与列表展示一致：status 为空且未转 MD 时界面为「正在提取」，也必须允许中断并写库
	if !dbg.BusyExtracting && !activeBefore {
		l.Infof("[cancel-convert] skip docId=%d: not busy and no active task", req.Id)
		return &types.Response{
			Code:    response.SuccessCode,
			Message: "当前文档未在识别中",
			Info:    l.debugJSON(dbg),
		}, nil
	}

	cancelledZip, cancelledMD, cancelledAny, cErr := CancelActiveConvertTasks(l.ctx, l.svcCtx, clientId, req.Id, "用户取消识别任务")
	if cErr != nil {
		l.Errorf("[cancel-convert] CancelActiveConvertTasks err=%v", cErr)
		return &types.Response{
			Code:    response.ServerErrorCode,
			Message: "取消任务失败",
			Info:    cErr.Error(),
		}, nil
	}
	dbg.CancelledZip = cancelledZip
	dbg.CancelledMD = cancelledMD

	revertStatus, rows, revErr := RevertRawDocumentAfterCancelExtract(l.ctx, l.svcCtx, clientId, req.Id)
	dbg.RevertTargetStatus = revertStatus
	dbg.RevertRowsAffected = rows
	if revErr != nil {
		l.Errorf("[cancel-convert] RevertRawDocumentAfterCancelExtract err=%v docId=%d", revErr, req.Id)
		return &types.Response{
			Code:    response.ServerErrorCode,
			Message: "恢复文档状态失败",
			Info:    revErr.Error() + "; " + l.debugJSON(dbg),
		}, nil
	}
	if rows == 0 {
		l.Errorf("[cancel-convert] revert rows=0 docId=%d clientId=%q docClientId=%q", req.Id, clientId, dbg.DocClientId)
		return &types.Response{
			Code:    response.ServerErrorCode,
			Message: "恢复文档状态失败：未更新任何记录",
			Info:    l.debugJSON(dbg),
		}, nil
	}

	asynctasksignal.BumpClientWatermark(l.ctx, l.svcCtx.RedisClient, clientId)

	if after, e := l.svcCtx.RawDocumentsModel.FindOneByClientId(l.ctx, clientId, req.Id); e == nil && after != nil {
		dbg.StatusAfter = strings.TrimSpace(after.Status)
	}

	l.Infof("[cancel-convert] done docId=%d cancelledAny=%v revert=%q rows=%d statusAfter=%q debug=%s",
		req.Id, cancelledAny, revertStatus, rows, dbg.StatusAfter, l.debugJSON(dbg))

	msg := "已恢复文档状态为「" + revertStatus + "」"
	if cancelledAny {
		msg = "已提交取消请求，" + msg
	} else {
		msg = "未发现进行中的识别任务（可能服务已重启），" + msg
	}

	return &types.Response{
		Code:    response.SuccessCode,
		Message: msg,
		Info:    l.debugJSON(dbg),
	}, nil
}

func (l *CancelConvertDocumentLogic) debugJSON(d cancelConvertDebug) string {
	b, err := json.Marshal(d)
	if err != nil {
		return fmt.Sprintf("%+v", d)
	}
	return string(b)
}

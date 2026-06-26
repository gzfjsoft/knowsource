// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package knowsource

import (
	"context"
	"strings"

	"knowsource/api/internal/svc"
	"knowsource/api/internal/types"
	"knowsource/common/response"
	"knowsource/model"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListAsyncTaskLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// list async tasks for current tenant
func NewListAsyncTaskLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListAsyncTaskLogic {
	return &ListAsyncTaskLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ListAsyncTaskLogic) ListAsyncTask(req *types.ListAsyncTaskRequest) (resp *types.ListAsyncTaskResponse, err error) {
	clientId, _ := l.ctx.Value("clientId").(string)
	clientId = strings.TrimSpace(clientId)
	if clientId == "" {
		return &types.ListAsyncTaskResponse{
			Response: types.Response{Code: response.UnauthorizedCode, Message: "clientId不能为空，请重新登录"},
		}, nil
	}

	page := int64(req.Page)
	pageSize := int64(req.PageSize)
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}
	if pageSize > 200 {
		pageSize = 200
	}
	offset := (page - 1) * pageSize
	// 空 status：查询全部状态（含 success）
	statusFilter := strings.TrimSpace(req.Status)

	taskModel := model.NewAsyncTaskModel(l.svcCtx.Mysql)
	rows, qErr := taskModel.FindByClientId(l.ctx, clientId, strings.TrimSpace(req.TaskType), statusFilter, offset, pageSize)
	if qErr != nil {
		return &types.ListAsyncTaskResponse{
			Response: types.Response{Code: response.ServerErrorCode, Message: "查询失败", Info: qErr.Error()},
		}, nil
	}
	total, cErr := taskModel.CountByClientId(l.ctx, clientId, strings.TrimSpace(req.TaskType), statusFilter)
	if cErr != nil {
		return &types.ListAsyncTaskResponse{
			Response: types.Response{Code: response.ServerErrorCode, Message: "查询失败", Info: cErr.Error()},
		}, nil
	}

	list := make([]types.AsyncTaskItem, 0, len(rows))
	for _, r := range rows {
		if r == nil {
			continue
		}
		list = append(list, types.AsyncTaskItem{
			Id:            r.Id,
			ClientId:      r.ClientId,
			TaskType:      r.TaskType,
			TaskDesc:      r.TaskDesc,
			SourceId:      r.SourceId,
			SourceKey:     r.SourceKey,
			ExecuteResult: r.ExecuteResult,
			Status:        r.Status,
			CreatedAt:     r.CreatedAt.Unix(),
			UpdatedAt:     r.UpdatedAt.Unix(),
		})
	}

	return &types.ListAsyncTaskResponse{
		Response: types.Response{Code: response.SuccessCode, Message: "success"},
		Data: &types.ListAsyncTaskData{
			List:  list,
			Total: total,
		},
	}, nil
}

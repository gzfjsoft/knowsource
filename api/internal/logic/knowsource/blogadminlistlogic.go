// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package knowsource

import (
	"context"
	"strings"

	"knowsource/api/internal/svc"
	"knowsource/api/internal/types"
	"knowsource/common/response"

	"github.com/zeromicro/go-zero/core/logx"
)

type BlogAdminListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 博客管理列表（仅 admin 租户 superadmin）
func NewBlogAdminListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *BlogAdminListLogic {
	return &BlogAdminListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *BlogAdminListLogic) BlogAdminList(req *types.BlogAdminListRequest) (resp *types.BlogListResponse, err error) {
	if denied := ensureBlogAdminPermission(l.ctx); denied != nil {
		return &types.BlogListResponse{Response: *denied}, nil
	}
	page := req.Page
	pageSize := req.PageSize
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	isPublished := req.IsPublished
	if isPublished != 0 && isPublished != 1 {
		isPublished = -1
	}
	rows, total, qErr := l.svcCtx.BlogModel.FindAdminList(l.ctx, strings.TrimSpace(req.Keyword), isPublished, int64(page), int64(pageSize))
	if qErr != nil {
		l.Errorf("BlogModel.FindAdminList err=%v", qErr)
		return &types.BlogListResponse{
			Response: types.Response{
				Code:    response.ServerErrorCode,
				Message: "查询博客列表失败",
				Info:    qErr.Error(),
			},
		}, nil
	}
	list := make([]types.BlogItem, 0, len(rows))
	for _, row := range rows {
		list = append(list, mapBlogRecordToItem(row, false))
	}
	return &types.BlogListResponse{
		Response: types.Response{
			Code:    response.SuccessCode,
			Message: "success",
		},
		Data: &types.BlogListData{
			List:  list,
			Total: total,
		},
	}, nil
}

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

type BlogListPublicLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 公开博客列表（无需登录）
func NewBlogListPublicLogic(ctx context.Context, svcCtx *svc.ServiceContext) *BlogListPublicLogic {
	return &BlogListPublicLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *BlogListPublicLogic) BlogListPublic(req *types.BlogListPublicRequest) (resp *types.BlogListResponse, err error) {
	page := req.Page
	pageSize := req.PageSize
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	rows, total, qErr := l.svcCtx.BlogModel.FindPublicList(l.ctx, strings.TrimSpace(req.Keyword), int64(page), int64(pageSize))
	if qErr != nil {
		l.Errorf("BlogModel.FindPublicList err=%v", qErr)
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

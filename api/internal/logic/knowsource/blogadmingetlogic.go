// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package knowsource

import (
	"context"

	"knowsource/api/internal/svc"
	"knowsource/api/internal/types"
	"knowsource/common/response"
	"knowsource/model"

	"github.com/zeromicro/go-zero/core/logx"
)

type BlogAdminGetLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 博客管理详情（仅 admin 租户 superadmin）
func NewBlogAdminGetLogic(ctx context.Context, svcCtx *svc.ServiceContext) *BlogAdminGetLogic {
	return &BlogAdminGetLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *BlogAdminGetLogic) BlogAdminGet(req *types.BlogAdminGetRequest) (resp *types.BlogReadResponse, err error) {
	if denied := ensureBlogAdminPermission(l.ctx); denied != nil {
		return &types.BlogReadResponse{Response: *denied}, nil
	}
	if req.Id <= 0 {
		return &types.BlogReadResponse{
			Response: types.Response{
				Code:    response.ParameterErrorCode,
				Message: "id 无效",
			},
		}, nil
	}
	row, qErr := l.svcCtx.BlogModel.FindAdminById(l.ctx, req.Id)
	if qErr != nil {
		if qErr == model.ErrNotFound {
			return &types.BlogReadResponse{
				Response: types.Response{
					Code:    response.RecordNotExistCode,
					Message: "博客不存在",
				},
			}, nil
		}
		l.Errorf("BlogModel.FindAdminById err=%v id=%d", qErr, req.Id)
		return &types.BlogReadResponse{
			Response: types.Response{
				Code:    response.ServerErrorCode,
				Message: "查询博客失败",
				Info:    qErr.Error(),
			},
		}, nil
	}
	item := mapBlogRecordToItem(row, true)
	return &types.BlogReadResponse{
		Response: types.Response{
			Code:    response.SuccessCode,
			Message: "success",
		},
		Data: &item,
	}, nil
}

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

type BlogReadLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 公开博客详情（通过 alias 或 id，无需登录）
func NewBlogReadLogic(ctx context.Context, svcCtx *svc.ServiceContext) *BlogReadLogic {
	return &BlogReadLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *BlogReadLogic) BlogRead(req *types.BlogReadRequest) (resp *types.BlogReadResponse, err error) {
	slug := normalizeBlogAlias(req.Slug)
	if slug == "" {
		return &types.BlogReadResponse{
			Response: types.Response{
				Code:    response.ParameterErrorCode,
				Message: "slug 不能为空",
			},
		}, nil
	}
	row, qErr := l.svcCtx.BlogModel.FindPublicBySlug(l.ctx, strings.TrimSpace(slug))
	if qErr != nil {
		if qErr == model.ErrNotFound {
			return &types.BlogReadResponse{
				Response: types.Response{
					Code:    response.RecordNotExistCode,
					Message: "博客不存在或未发布",
				},
			}, nil
		}
		l.Errorf("BlogModel.FindPublicBySlug err=%v slug=%s", qErr, slug)
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

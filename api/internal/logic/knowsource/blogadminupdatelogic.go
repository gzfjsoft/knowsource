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

type BlogAdminUpdateLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 博客管理更新（仅 admin 租户 superadmin）
func NewBlogAdminUpdateLogic(ctx context.Context, svcCtx *svc.ServiceContext) *BlogAdminUpdateLogic {
	return &BlogAdminUpdateLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *BlogAdminUpdateLogic) BlogAdminUpdate(req *types.BlogAdminUpdateRequest) (resp *types.Response, err error) {
	if denied := ensureBlogAdminPermission(l.ctx); denied != nil {
		return denied, nil
	}
	if req.Id <= 0 || strings.TrimSpace(req.Title) == "" || strings.TrimSpace(req.Content) == "" {
		return &types.Response{
			Code:    response.ParameterErrorCode,
			Message: "参数无效，标题和内容不能为空",
		}, nil
	}
	if _, fErr := l.svcCtx.BlogModel.FindAdminById(l.ctx, req.Id); fErr != nil {
		if fErr == model.ErrNotFound {
			return &types.Response{Code: response.RecordNotExistCode, Message: "博客不存在"}, nil
		}
		return &types.Response{Code: response.ServerErrorCode, Message: "查询博客失败", Info: fErr.Error()}, nil
	}
	alias := normalizeBlogAlias(req.Alias)
	if ok, e := l.svcCtx.BlogModel.ExistsAlias(l.ctx, alias, req.Id); e != nil {
		return &types.Response{Code: response.ServerErrorCode, Message: "校验别名失败", Info: e.Error()}, nil
	} else if ok {
		return &types.Response{Code: response.ConflictCode, Message: "别名已存在"}, nil
	}
	isPublished := req.IsPublished
	if isPublished != 1 {
		isPublished = 0
	}
	uErr := l.svcCtx.BlogModel.UpdateAdmin(l.ctx, &model.BlogAdminMutation{
		Id:          req.Id,
		Title:       req.Title,
		Alias:       alias,
		Summary:     req.Summary,
		Content:     req.Content,
		Tags:        req.Tags,
		Categories:  req.Categories,
		Authors:     req.Authors,
		Banner:      req.Banner,
		IsPublished: isPublished,
	})
	if uErr != nil {
		l.Errorf("BlogModel.UpdateAdmin err=%v id=%d", uErr, req.Id)
		return &types.Response{
			Code:    response.ServerErrorCode,
			Message: "更新博客失败",
			Info:    uErr.Error(),
		}, nil
	}
	return &types.Response{
		Code:    response.SuccessCode,
		Message: "success",
	}, nil
}

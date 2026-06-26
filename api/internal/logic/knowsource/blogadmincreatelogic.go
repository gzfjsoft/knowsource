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

type BlogAdminCreateLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 博客管理创建（仅 admin 租户 superadmin）
func NewBlogAdminCreateLogic(ctx context.Context, svcCtx *svc.ServiceContext) *BlogAdminCreateLogic {
	return &BlogAdminCreateLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *BlogAdminCreateLogic) BlogAdminCreate(req *types.BlogAdminCreateRequest) (resp *types.Response, err error) {
	if denied := ensureBlogAdminPermission(l.ctx); denied != nil {
		return denied, nil
	}
	if strings.TrimSpace(req.Title) == "" || strings.TrimSpace(req.Content) == "" {
		return &types.Response{
			Code:    response.ParameterErrorCode,
			Message: "标题和内容不能为空",
		}, nil
	}
	alias := normalizeBlogAlias(req.Alias)
	if ok, e := l.svcCtx.BlogModel.ExistsAlias(l.ctx, alias, 0); e != nil {
		return &types.Response{Code: response.ServerErrorCode, Message: "校验别名失败", Info: e.Error()}, nil
	} else if ok {
		return &types.Response{Code: response.ConflictCode, Message: "别名已存在"}, nil
	}
	isPublished := req.IsPublished
	if isPublished != 1 {
		isPublished = 0
	}
	_, cErr := l.svcCtx.BlogModel.CreateAdmin(l.ctx, &model.BlogAdminMutation{
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
	if cErr != nil {
		l.Errorf("BlogModel.CreateAdmin err=%v", cErr)
		return &types.Response{
			Code:    response.ServerErrorCode,
			Message: "创建博客失败",
			Info:    cErr.Error(),
		}, nil
	}
	return &types.Response{
		Code:    response.SuccessCode,
		Message: "success",
	}, nil
}

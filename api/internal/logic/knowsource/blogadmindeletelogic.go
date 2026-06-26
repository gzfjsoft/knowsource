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

type BlogAdminDeleteLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 博客管理删除（仅 admin 租户 superadmin）
func NewBlogAdminDeleteLogic(ctx context.Context, svcCtx *svc.ServiceContext) *BlogAdminDeleteLogic {
	return &BlogAdminDeleteLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *BlogAdminDeleteLogic) BlogAdminDelete(req *types.BlogAdminDeleteRequest) (resp *types.Response, err error) {
	if denied := ensureBlogAdminPermission(l.ctx); denied != nil {
		return denied, nil
	}
	if req.Id <= 0 {
		return &types.Response{
			Code:    response.ParameterErrorCode,
			Message: "id 无效",
		}, nil
	}
	if _, fErr := l.svcCtx.BlogModel.FindAdminById(l.ctx, req.Id); fErr != nil {
		if fErr == model.ErrNotFound {
			return &types.Response{Code: response.RecordNotExistCode, Message: "博客不存在"}, nil
		}
		return &types.Response{Code: response.ServerErrorCode, Message: "查询博客失败", Info: fErr.Error()}, nil
	}
	if dErr := l.svcCtx.BlogModel.DeleteAdmin(l.ctx, req.Id); dErr != nil {
		l.Errorf("BlogModel.DeleteAdmin err=%v id=%d", dErr, req.Id)
		return &types.Response{
			Code:    response.ServerErrorCode,
			Message: "删除博客失败",
			Info:    dErr.Error(),
		}, nil
	}
	return &types.Response{
		Code:    response.SuccessCode,
		Message: "success",
	}, nil
}

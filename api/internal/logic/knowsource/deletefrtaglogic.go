package knowsource

import (
	"context"
	"strings"

	"knowsource/api/internal/svc"
	"knowsource/api/internal/types"
	"knowsource/common/response"
	"knowsource/model"

	"github.com/pkg/errors"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type DeleteFrTagLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 删除标签
func NewDeleteFrTagLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteFrTagLogic {
	return &DeleteFrTagLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *DeleteFrTagLogic) DeleteFrTag(req *types.FrTagDeleteRequest) (resp *types.Response, err error) {
	clientId, _ := l.ctx.Value("clientId").(string)
	clientId = strings.TrimSpace(clientId)
	if clientId == "" {
		return &types.Response{
			Code:    response.UnauthorizedCode,
			Message: "clientId不能为空，请重新登录",
		}, nil
	}

	if req.Tag == "" {
		return &types.Response{
			Code:    response.ParameterErrorCode,
			Message: "标签名称不能为空",
		}, nil
	}

	// 检查标签是否存在
	_, err = l.svcCtx.FrTagsModel.FindOneByClientIdTag(l.ctx, clientId, req.Tag)
	if err != nil {
		if errors.Is(err, model.ErrNotFound) || err == sqlx.ErrNotFound {
			return &types.Response{
				Code:    response.RecordNotExistCode,
				Message: "标签不存在",
			}, nil
		}
		l.Logger.Errorf("查询标签失败: %v, Tag: %s", err, req.Tag)
		return &types.Response{
			Code:    response.ServerErrorCode,
			Message: "Database error",
			Info:    err.Error(),
		}, nil
	}

	// 若有 raw_doc 使用该标签（无论是否已审核），则不允许删除
	count, err := l.svcCtx.RawDocumentsModel.CountAuditedByTag(l.ctx, clientId, req.Tag)
	if err != nil {
		l.Logger.Errorf("检查 raw_documents 标签使用情况失败: %v, Tag: %s", err, req.Tag)
		return &types.Response{
			Code:    response.ServerErrorCode,
			Message: "检查失败",
			Info:    err.Error(),
		}, nil
	}
	if count > 0 {
		return &types.Response{
			Code:    response.ConflictCode,
			Message: "存在文档使用该标签，不能删除；请先更改文档标签后再删除",
		}, nil
	}

	// 删除标签
	err = l.svcCtx.FrTagsModel.DeleteByClientIdTag(l.ctx, clientId, req.Tag)
	if err != nil {
		l.Logger.Errorf("删除标签失败: %v, Tag: %s", err, req.Tag)
		return &types.Response{
			Code:    response.ServerErrorCode,
			Message: "Database error",
			Info:    err.Error(),
		}, nil
	}

	return &types.Response{
		Code:    response.SuccessCode,
		Message: "success",
	}, nil
}

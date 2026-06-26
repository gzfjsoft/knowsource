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

type UpdateFrTagLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 更新标签
func NewUpdateFrTagLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateFrTagLogic {
	return &UpdateFrTagLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UpdateFrTagLogic) UpdateFrTag(req *types.FrTagUpdateRequest) (resp *types.Response, err error) {
	clientId, _ := l.ctx.Value("clientId").(string)
	clientId = strings.TrimSpace(clientId)
	if clientId == "" {
		return &types.Response{
			Code:    response.UnauthorizedCode,
			Message: "clientId不能为空，请重新登录",
		}, nil
	}

	if req.OldTag == "" {
		return &types.Response{
			Code:    response.ParameterErrorCode,
			Message: "旧标签名称不能为空",
		}, nil
	}

	if req.NewTag == "" {
		return &types.Response{
			Code:    response.ParameterErrorCode,
			Message: "新标签名称不能为空",
		}, nil
	}

	// 如果新旧标签相同，直接返回成功
	if req.OldTag == req.NewTag {
		return &types.Response{
			Code:    response.SuccessCode,
			Message: "success",
		}, nil
	}

	// 检查旧标签是否存在
	_, err = l.svcCtx.FrTagsModel.FindOneByClientIdTag(l.ctx, clientId, req.OldTag)
	if err != nil {
		if errors.Is(err, model.ErrNotFound) || err == sqlx.ErrNotFound {
			return &types.Response{
				Code:    response.RecordNotExistCode,
				Message: "旧标签不存在",
			}, nil
		}
		l.Logger.Errorf("查询旧标签失败: %v, OldTag: %s", err, req.OldTag)
		return &types.Response{
			Code:    response.ServerErrorCode,
			Message: "Database error",
			Info:    err.Error(),
		}, nil
	}

	// 检查新标签是否已存在
	_, err = l.svcCtx.FrTagsModel.FindOneByClientIdTag(l.ctx, clientId, req.NewTag)
	if err == nil {
		return &types.Response{
			Code:    response.ParameterErrorCode,
			Message: "新标签已存在",
		}, nil
	}

	if !errors.Is(err, model.ErrNotFound) && err != sqlx.ErrNotFound {
		l.Logger.Errorf("查询新标签失败: %v, NewTag: %s", err, req.NewTag)
		return &types.Response{
			Code:    response.ServerErrorCode,
			Message: "Database error",
			Info:    err.Error(),
		}, nil
	}

	err = l.svcCtx.FrTagsModel.UpdateTagByClientId(l.ctx, clientId, req.OldTag, req.NewTag)
	if err != nil {
		l.Logger.Errorf("更新标签失败: %v, OldTag: %s, NewTag: %s", err, req.OldTag, req.NewTag)
		return &types.Response{
			Code:    response.ServerErrorCode,
			Message: "更新标签失败",
			Info:    err.Error(),
		}, nil
	}

	return &types.Response{
		Code:    response.SuccessCode,
		Message: "success",
	}, nil
}

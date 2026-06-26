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
)

type CreateFrTagLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 创建标签
func NewCreateFrTagLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateFrTagLogic {
	return &CreateFrTagLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CreateFrTagLogic) CreateFrTag(req *types.FrTagCreateRequest) (resp *types.Response, err error) {
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

	// 检查是否已存在相同的标签
	_, err = l.svcCtx.FrTagsModel.FindOneByClientIdTag(l.ctx, clientId, req.Tag)
	if err == nil {
		return &types.Response{
			Code:    response.ParameterErrorCode,
			Message: "该标签已存在",
		}, nil
	}

	if !errors.Is(err, model.ErrNotFound) {
		l.Logger.Errorf("查询标签失败: %v", err)
		return &types.Response{
			Code:    response.ServerErrorCode,
			Message: "Database error",
			Info:    err.Error(),
		}, nil
	}

	// 创建新标签
	tag := &model.Tags{
		ClientId: clientId,
		Tag:      req.Tag,
	}

	_, err = l.svcCtx.FrTagsModel.Insert(l.ctx, tag)
	if err != nil {
		l.Logger.Errorf("创建标签失败: %v", err)
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

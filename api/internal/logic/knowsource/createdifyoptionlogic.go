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

type CreateDifyOptionLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// get dify option
func NewCreateDifyOptionLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateDifyOptionLogic {
	return &CreateDifyOptionLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CreateDifyOptionLogic) CreateDifyOption(req *types.GetDifyOptionData) (resp *types.Response, err error) {
	clientId, _ := l.ctx.Value("clientId").(string)
	clientId = strings.TrimSpace(clientId)
	if clientId == "" {
		return &types.Response{Code: response.UnauthorizedCode, Message: "clientId不能为空，请重新登录"}, nil
	}

	if req.Name == "" {
		return &types.Response{
			Code:    response.ParameterErrorCode,
			Message: "name参数不能为空",
		}, nil
	}

	// 检查是否已存在
	_, err = l.svcCtx.DifyOptionsModel.FindOneByClientIdName(l.ctx, clientId, req.Name)
	if err == nil {
		// 记录已存在，执行更新
		option := &model.DifyOptions{
			ClientId: clientId,
			Name:     req.Name,
		}

		option.Url = req.Url

		option.ApiKey = req.ApiKey

		option.Description = req.Description

		err = l.svcCtx.DifyOptionsModel.Update(l.ctx, option)
		if err != nil {
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

	if !errors.Is(err, model.ErrNotFound) {
		return &types.Response{
			Code:    response.ServerErrorCode,
			Message: "Database error",
			Info:    err.Error(),
		}, nil
	}

	// 记录不存在，执行插入
	option := &model.DifyOptions{
		ClientId:    clientId,
		Name:        req.Name,
		Description: req.Description,
		Url:         req.Url,
		ApiKey:      req.ApiKey,
	}

	_, err = l.svcCtx.DifyOptionsModel.Insert(l.ctx, option)
	if err != nil {
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

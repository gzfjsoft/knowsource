package knowdata

import (
	"context"
	"knowsource/common/response"
	"knowsource/model"

	"knowsource/api/internal/svc"
	"knowsource/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type CreateAIConfigLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewCreateAIConfigLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateAIConfigLogic {
	return &CreateAIConfigLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CreateAIConfigLogic) CreateAIConfig(req *types.KnowdataCreateAIConfigRequest) (resp *response.Response, err error) {
	userName, _ := l.ctx.Value("userName").(string)
	aiConfig := &model.AiConfig{
		DocumentCode: req.DocumentCode,
		Name:         req.Name,
		Value:        req.Value,

		CreatedBy: userName,
	}
	_, err = l.svcCtx.AiConfigModel.Insert(l.ctx, aiConfig)
	if err != nil {
		l.Logger.Errorf("创建AI配置失败: %v", err)
		return &response.Response{
			Code:    response.ServerErrorCode,
			Message: "创建AI配置失败，请稍后重试",
		}, nil
	}

	return &response.Response{
		Code:    response.SuccessCode,
		Message: "success",
	}, nil
}

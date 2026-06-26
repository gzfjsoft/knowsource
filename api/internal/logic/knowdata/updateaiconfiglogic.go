package knowdata

import (
	"context"
	"database/sql"
	"knowsource/common/response"
	"knowsource/model"
	"strings"

	"github.com/pkg/errors"

	"knowsource/api/internal/svc"
	"knowsource/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateAIConfigLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUpdateAIConfigLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateAIConfigLogic {
	return &UpdateAIConfigLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UpdateAIConfigLogic) UpdateAIConfig(req *types.KnowdataUpdateAIConfigRequest) (resp *response.Response, err error) {
	clientId, _ := l.ctx.Value("clientId").(string)
	if strings.TrimSpace(clientId) == "" {
		return &response.Response{
			Code:    response.UnauthorizedCode,
			Message: "clientId不能为空，请重新登录",
		}, nil
	}

	res, err := l.svcCtx.AiConfigModel.FindOne(l.ctx, req.Id)
	if err != nil {
		if errors.Is(err, model.ErrNotFound) {
			// 记录不存在的情况
			l.Logger.Info("AI配置记录不存在", logx.Field("id", req.Id))
			return &response.Response{
				Code:    response.RecordNotExistCode,
				Message: "记录不存在或已删除",
			}, nil
		}
		// 其他查询错误
		l.Logger.Error("查询AI配置失败", logx.Field("id", req.Id), logx.Field("error", err))
		return &response.Response{
			Code:    response.ServerErrorCode,
			Message: err.Error(),
		}, nil
	}

	userName, ok := l.ctx.Value("userName").(string)
	if !ok || userName == "" {
		userName = "unknown" // 默认用户名
	}

	aiConfig := &model.AiConfig{
		Id:           req.Id,
		ClientId:     res.ClientId,
		Name:         req.Name,
		DocumentCode: req.DocumentCode,
		Value:        req.Value,
		CreatedBy:    res.CreatedBy,
		CreatedAt:    res.CreatedAt,
		UpdatedBy:    sql.NullString{String: userName, Valid: strings.TrimSpace(userName) != ""},
	}
	err = l.svcCtx.AiConfigModel.Update(l.ctx, aiConfig)
	if err != nil {
		l.Logger.Errorf("更新视频失败: %v", err)
		return &response.Response{
			Code:    response.ServerErrorCode,
			Message: "更新AI配置失败，请稍后重试",
		}, nil
	}

	return &response.Response{
		Code:    response.SuccessCode,
		Message: "success",
	}, nil
}

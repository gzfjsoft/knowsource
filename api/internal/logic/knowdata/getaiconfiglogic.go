package knowdata

import (
	"context"
	"knowsource/common/response"
	"knowsource/model"
	"strings"

	"github.com/pkg/errors"

	"knowsource/api/internal/svc"
	"knowsource/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetAIConfigLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetAIConfigLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetAIConfigLogic {
	return &GetAIConfigLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetAIConfigLogic) GetAIConfig(req *types.PathIdRequest) (resp *response.Response, err error) {
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

	// 转换数据格式
	updatedBy := ""
	if res.UpdatedBy.Valid {
		updatedBy = res.UpdatedBy.String
	}
	aiConfigInfo := &types.KnowdataAIConfigInfo{
		Id:           res.Id,
		Name:         res.Name,
		DocumentCode: res.DocumentCode,
		Value:        res.Value,
		CreatedAt:    res.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt:    res.UpdatedAt.Format("2006-01-02 15:04:05"),
		CreatedBy:    res.CreatedBy,
		UpdatedBy:    updatedBy,
	}

	return &response.Response{
		Code:    response.SuccessCode,
		Message: "success",
		Data:    aiConfigInfo,
	}, nil
}

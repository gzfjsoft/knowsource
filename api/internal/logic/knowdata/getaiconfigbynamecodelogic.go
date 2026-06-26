package knowdata

import (
	"context"
	"errors"
	"strings"

	"knowsource/api/internal/svc"
	"knowsource/api/internal/types"
	"knowsource/common/response"
	"knowsource/model"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetAIConfigByNameCodeLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 获取AI配置byname和documentCode
func NewGetAIConfigByNameCodeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetAIConfigByNameCodeLogic {
	return &GetAIConfigByNameCodeLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetAIConfigByNameCodeLogic) GetAIConfigByNameCode(req *types.GetAIConfigByNameCodeRequest) (resp *types.GetAIConfigByNameResp, err error) {
	clientId, _ := l.ctx.Value("clientId").(string)
	if strings.TrimSpace(clientId) == "" {
		return &types.GetAIConfigByNameResp{
			Response: types.Response{
				Code:    response.UnauthorizedCode,
				Message: "clientId不能为空，请重新登录",
			},
			Data: nil,
		}, nil
	}

	res, err := l.svcCtx.AiConfigModel.FindOneByClientIdDocumentCodeName(l.ctx, clientId, req.DocumentCode, req.Name)
	if err != nil {
		if errors.Is(err, model.ErrNotFound) {
			// 记录不存在的情况
			l.Logger.Info("AI配置记录不存在", logx.Field("name", req.Name), logx.Field("documentCode", req.DocumentCode))
			return &types.GetAIConfigByNameResp{
				Response: types.Response{
					Code:    response.RecordNotExistCode,
					Message: "记录不存在或已删除",
				},
				Data: nil,
			}, nil
		}
		// 其他查询错误
		l.Logger.Error("查询AI配置失败", logx.Field("name", req.Name), logx.Field("documentCode", req.DocumentCode), logx.Field("error", err))
		return &types.GetAIConfigByNameResp{
			Response: types.Response{
				Code:    response.ServerErrorCode,
				Message: err.Error(),
			},
			Data: nil,
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
		Value:        res.Value,
		CreatedAt:    res.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt:    res.UpdatedAt.Format("2006-01-02 15:04:05"),
		CreatedBy:    res.CreatedBy,
		UpdatedBy:    updatedBy,
		DocumentCode: req.DocumentCode, // 使用请求中的 documentCode
	}

	return &types.GetAIConfigByNameResp{
		Response: types.Response{
			Code:    response.SuccessCode,
			Message: "success",
		},
		Data: aiConfigInfo,
	}, nil
}

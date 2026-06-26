package knowdata

import (
	"context"
	"knowsource/api/internal/utils"
	"knowsource/common/response"
	"knowsource/model"
	"strings"

	"knowsource/api/internal/svc"
	"knowsource/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetAIConfigListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetAIConfigListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetAIConfigListLogic {
	return &GetAIConfigListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetAIConfigListLogic) GetAIConfigList(req *types.KnowdataAIConfigListRequest) (resp *response.Response, err error) {
	// 设置默认分页参数（form 标签解析 GET 查询参数 page/pageSize）
	intPage, intPageSize := utils.FixPagingParam(uint64(req.Page), uint64(req.PageSize))
	req.Page = int64(intPage)
	req.PageSize = int64(intPageSize)
	page := uint64(req.Page)
	pageSize := uint64(req.PageSize)

	clientId, _ := l.ctx.Value("clientId").(string)
	if strings.TrimSpace(clientId) == "" {
		return &response.Response{
			Code:    response.UnauthorizedCode,
			Message: "clientId不能为空，请重新登录",
		}, nil
	}

	var list []*model.AiConfig
	var total int64
	if req.DocumentCode != "" {
		// 使用支持 documentCode 的查询方法
		list, total, err = l.svcCtx.AiConfigModel.FindListWithDocumentCode(l.ctx, clientId, req.Name, req.DocumentCode, page, pageSize)
	} else {
		// 使用原有的查询方法
		list, total, err = l.svcCtx.AiConfigModel.FindListOrderByDocumentCode(l.ctx, clientId, req.Name, page, pageSize)
	}
	if err != nil {
		return &response.Response{
			Code:    response.ServerErrorCode,
			Message: err.Error(),
		}, nil
	}
	if total == 0 {
		return &response.Response{
			Code:    response.SuccessCode,
			Message: "ok",
			Data: types.PageResponse{
				Total: 0,
				List:  []interface{}{},
			},
		}, nil
	}

	// 转换数据格式
	result := make([]types.KnowdataAIConfigInfo, 0, len(list))
	for _, item := range list {
		updatedBy := ""
		if item.UpdatedBy.Valid {
			updatedBy = item.UpdatedBy.String
		}
		result = append(result, types.KnowdataAIConfigInfo{
			Id:           item.Id,
			Name:         item.Name,
			Value:        item.Value,
			DocumentCode: item.DocumentCode,
			CreatedAt:    item.CreatedAt.Format("2006-01-02 15:04:05"),
			UpdatedAt:    item.UpdatedAt.Format("2006-01-02 15:04:05"),
			CreatedBy:    item.CreatedBy,
			UpdatedBy:    updatedBy,
		})
	}

	return &response.Response{
		Code:    response.SuccessCode,
		Message: "ok",
		Data: types.PageResponse{
			Total: total,
			List:  result,
		},
	}, nil
}

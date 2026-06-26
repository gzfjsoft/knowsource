package knowdata

import (
	"context"
	"strings"

	"knowsource/api/internal/svc"
	"knowsource/api/internal/types"
	"knowsource/common/response"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetDistinctTagsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 获取所有不重复的Tag标签
func NewGetDistinctTagsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetDistinctTagsLogic {
	return &GetDistinctTagsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetDistinctTagsLogic) GetDistinctTags(req *types.GetDistinctTagsRequest) (resp *types.GetDistinctTagsResponse, err error) {
	clientId, _ := l.ctx.Value("clientId").(string)
	clientId = strings.TrimSpace(clientId)
	if clientId == "" {
		return &types.GetDistinctTagsResponse{
			Response: types.Response{
				Code:    response.UnauthorizedCode,
				Message: "clientId不能为空，请重新登录",
			},
		}, nil
	}

	resp = &types.GetDistinctTagsResponse{
		Response: types.Response{
			Code:    response.SuccessCode,
			Message: "获取成功",
		},
	}

	tags, err := l.svcCtx.RawDocumentsModel.FindDistinctTags(l.ctx, clientId, req.DocumentCode)
	if err != nil {
		l.Logger.Errorf("获取Tag列表失败: %v", err)
		resp.Response.Code = response.ServerErrorCode
		resp.Response.Message = "获取失败"
		resp.Response.Info = err.Error()
		return resp, nil
	}

	resp.Data = types.GetDistinctTagsData{
		Tags: tags,
	}

	return resp, nil
}

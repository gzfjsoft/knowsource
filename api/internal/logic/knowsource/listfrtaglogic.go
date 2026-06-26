package knowsource

import (
	"context"
	"strings"

	"knowsource/api/internal/svc"
	"knowsource/api/internal/types"
	"knowsource/common/response"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type ListFrTagLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 获取标签列表
func NewListFrTagLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListFrTagLogic {
	return &ListFrTagLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ListFrTagLogic) ListFrTag(req *types.FrTagListRequest) (resp *types.FrTagListResponse, err error) {
	clientId, _ := l.ctx.Value("clientId").(string)
	clientId = strings.TrimSpace(clientId)
	if clientId == "" {
		return &types.FrTagListResponse{
			Response: types.Response{
				Code:    response.UnauthorizedCode,
				Message: "clientId不能为空，请重新登录",
			},
		}, nil
	}

	// 设置默认分页参数
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 10
	}

	// 查询总数
	total, err := l.svcCtx.FrTagsModel.CountWithCondition(l.ctx, clientId, req.Tag)
	if err != nil {
		l.Logger.Errorf("获取标签总数失败: %v", err)
		return &types.FrTagListResponse{
			Response: types.Response{
				Code:    response.ServerErrorCode,
				Message: "获取标签总数失败",
				Info:    err.Error(),
			},
		}, nil
	}

	// 查询列表数据
	offset := int64((req.Page - 1) * req.PageSize)
	pageSize := int64(req.PageSize)
	rows, err := l.svcCtx.FrTagsModel.FindAllWithCondition(l.ctx, clientId, req.Tag, offset, pageSize)
	if err != nil {
		if err == sqlx.ErrNotFound {
			return &types.FrTagListResponse{
				Response: types.Response{
					Code:    response.SuccessCode,
					Message: "查询成功",
				},
				Data: &types.FrTagListData{
					List:  []types.FrTagInfo{},
					Total: 0,
				},
			}, nil
		}
		l.Logger.Errorf("获取标签列表失败: %v", err)
		return &types.FrTagListResponse{
			Response: types.Response{
				Code:    response.ServerErrorCode,
				Message: "获取标签列表失败",
				Info:    err.Error(),
			},
		}, nil
	}

	// 转换为响应格式
	list := make([]types.FrTagInfo, len(rows))
	for i, row := range rows {
		list[i] = types.FrTagInfo{
			Tag: row.Tag,
		}
	}

	return &types.FrTagListResponse{
		Response: types.Response{
			Code:    response.SuccessCode,
			Message: "success",
		},
		Data: &types.FrTagListData{
			List:  list,
			Total: total,
		},
	}, nil
}

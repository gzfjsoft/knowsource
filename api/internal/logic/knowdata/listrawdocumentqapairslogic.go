// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package knowdata

import (
	"context"
	"strings"

	"knowsource/api/internal/svc"
	"knowsource/api/internal/types"
	"knowsource/common/response"
	"knowsource/model"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type ListRawDocumentQaPairsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 查看审核入库时抽取的问答队列
func NewListRawDocumentQaPairsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListRawDocumentQaPairsLogic {
	return &ListRawDocumentQaPairsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ListRawDocumentQaPairsLogic) ListRawDocumentQaPairs(req *types.ListRawDocumentQaPairsRequest) (resp *types.ListRawDocumentQaPairsResponse, err error) {
	clientId, _ := l.ctx.Value("clientId").(string)
	clientId = strings.TrimSpace(clientId)
	if clientId == "" {
		return &types.ListRawDocumentQaPairsResponse{
			Response: types.Response{
				Code:    response.UnauthorizedCode,
				Message: "clientId不能为空，请重新登录",
			},
		}, nil
	}
	if req.Id <= 0 {
		return &types.ListRawDocumentQaPairsResponse{
			Response: types.Response{
				Code:    response.ParameterErrorCode,
				Message: "文档ID不能为空",
			},
		}, nil
	}

	if _, fErr := l.svcCtx.RawDocumentsModel.FindOneByClientId(l.ctx, clientId, req.Id); fErr != nil {
		if fErr == sqlx.ErrNotFound || fErr == model.ErrNotFound {
			return &types.ListRawDocumentQaPairsResponse{
				Response: types.Response{
					Code:    response.RecordNotExistCode,
					Message: "文档不存在",
				},
			}, nil
		}
		return &types.ListRawDocumentQaPairsResponse{
			Response: types.Response{
				Code:    response.ServerErrorCode,
				Message: "查询文档失败",
				Info:    fErr.Error(),
			},
		}, nil
	}

	page := int64(req.Page)
	pageSize := int64(req.PageSize)
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}
	offset := (page - 1) * pageSize

	total, cErr := l.svcCtx.RawDocumentQaPairsModel.CountByRawDocumentId(l.ctx, clientId, req.Id)
	if cErr != nil {
		return &types.ListRawDocumentQaPairsResponse{
			Response: types.Response{
				Code:    response.ServerErrorCode,
				Message: "统计问答失败",
				Info:    cErr.Error(),
			},
		}, nil
	}
	rows, qErr := l.svcCtx.RawDocumentQaPairsModel.ListByRawDocumentId(l.ctx, clientId, req.Id, offset, pageSize)
	if qErr != nil {
		return &types.ListRawDocumentQaPairsResponse{
			Response: types.Response{
				Code:    response.ServerErrorCode,
				Message: "查询问答失败",
				Info:    qErr.Error(),
			},
		}, nil
	}

	list := make([]types.RawDocumentQaPairItem, 0, len(rows))
	for _, item := range rows {
		if item == nil {
			continue
		}
		list = append(list, types.RawDocumentQaPairItem{
			Id:            item.Id,
			ChunkIndex:    item.ChunkIndex,
			Question:      item.Question,
			Answer:        item.Answer,
			QdrantPointId: item.QdrantPointId,
			CreatedAt:     item.CreatedAt.Unix(),
		})
	}

	return &types.ListRawDocumentQaPairsResponse{
		Response: types.Response{
			Code:    response.SuccessCode,
			Message: "success",
		},
		Data: &types.ListRawDocumentQaPairsData{
			List:  list,
			Total: total,
		},
	}, nil
}

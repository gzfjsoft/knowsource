package knowdata

import (
	"context"
	"strings"

	"knowsource/api/internal/svc"
	"knowsource/api/internal/types"
	"knowsource/common/response"
	"knowsource/model"

	"github.com/pkg/errors"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type GetRawDocumentsContentDiffLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 获取原始文档的原始内容和当前内容用于比较
func NewGetRawDocumentsContentDiffLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetRawDocumentsContentDiffLogic {
	return &GetRawDocumentsContentDiffLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetRawDocumentsContentDiffLogic) GetRawDocumentsContentDiff(req *types.GetRawDocumentsContentDiffRequest) (resp *types.GetRawDocumentsContentDiffResponse, err error) {
	clientId, _ := l.ctx.Value("clientId").(string)
	clientId = strings.TrimSpace(clientId)
	if clientId == "" {
		return &types.GetRawDocumentsContentDiffResponse{
			Response: types.Response{
				Code:    response.UnauthorizedCode,
				Message: "clientId不能为空，请重新登录",
			},
		}, nil
	}

	// 检查 ID 是否有效
	if req.Id <= 0 {
		return &types.GetRawDocumentsContentDiffResponse{
			Response: types.Response{
				Code:    response.ParameterErrorCode,
				Message: "ID 不能为空或无效",
			},
		}, nil
	}

	// 查询文档
	doc, err := l.svcCtx.RawDocumentsModel.FindOneByClientId(l.ctx, clientId, req.Id)
	if err != nil {
		if err == sqlx.ErrNotFound || errors.Is(err, model.ErrNotFound) {
			return &types.GetRawDocumentsContentDiffResponse{
				Response: types.Response{
					Code:    response.RecordNotExistCode,
					Message: "文档不存在",
				},
			}, nil
		}
		l.Logger.Errorf("查询文档失败: %v, ID: %d", err, req.Id)
		return &types.GetRawDocumentsContentDiffResponse{
			Response: types.Response{
				Code:    response.ServerErrorCode,
				Message: "查询失败",
				Info:    err.Error(),
			},
		}, nil
	}

	return &types.GetRawDocumentsContentDiffResponse{
		Response: types.Response{
			Code:    response.SuccessCode,
			Message: "获取成功",
		},
		Data: &types.GetRawDocumentsContentDiffData{
			Id:         doc.Id,
			FileName:   doc.FileName,
			ContentOrg: doc.ContentOrg,
			Content:    doc.Content,
		},
	}, nil
}

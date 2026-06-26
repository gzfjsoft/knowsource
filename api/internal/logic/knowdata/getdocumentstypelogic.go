package knowdata

import (
	"context"
	"strings"

	"knowsource/api/internal/svc"
	"knowsource/api/internal/types"
	"knowsource/common/response"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetDocumentsTypeLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 获取文档类型详情
func NewGetDocumentsTypeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetDocumentsTypeLogic {
	return &GetDocumentsTypeLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetDocumentsTypeLogic) GetDocumentsType(req *types.DocumentsRequest) (resp *types.GetDocumentsTypeResponse, err error) {
	clientId, _ := l.ctx.Value("clientId").(string)
	clientId = strings.TrimSpace(clientId)
	if clientId == "" {
		return &types.GetDocumentsTypeResponse{
			Response: types.Response{Code: response.UnauthorizedCode, Message: "clientId不能为空，请重新登录"},
		}, nil
	}

	docType, err := l.svcCtx.DocumentTypeModel.FindOneByClientIdCode(l.ctx, clientId, req.Code)
	if err != nil {
		resp = &types.GetDocumentsTypeResponse{
			Response: types.Response{
				Code:    response.NotFoundCode,
				Message: "文档类型不存在",
				Info:    err.Error(),
			},
			Data: nil,
		}
		return resp, nil
	}

	data := types.DocumentsType{
		Code:        docType.Code,
		Name:        docType.Name,
		IsDisabled:  docType.IsDisabled,
		Description: docType.Description,
		CreatedAt:   docType.CreatedAt.Unix(),
		UpdatedAt:   docType.UpdatedAt.Unix(),
	}
	resp = &types.GetDocumentsTypeResponse{
		Response: types.Response{
			Code:    response.SuccessCode,
			Message: "获取成功",
		},
		Data: &data,
	}
	return resp, nil
}

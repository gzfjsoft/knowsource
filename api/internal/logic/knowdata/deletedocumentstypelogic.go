package knowdata

import (
	"context"
	"strings"

	"knowsource/api/internal/svc"
	"knowsource/api/internal/types"
	"knowsource/common/response"

	"github.com/zeromicro/go-zero/core/logx"
)

type DeleteDocumentsTypeLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 删除文档类型
func NewDeleteDocumentsTypeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteDocumentsTypeLogic {
	return &DeleteDocumentsTypeLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *DeleteDocumentsTypeLogic) DeleteDocumentsType(req *types.DocumentsRequest) (resp *types.Response, err error) {
	clientId, _ := l.ctx.Value("clientId").(string)
	clientId = strings.TrimSpace(clientId)
	if clientId == "" {
		return &types.Response{
			Code:    response.UnauthorizedCode,
			Message: "clientId不能为空，请重新登录",
		}, nil
	}

	// Check if document type exists
	_, err = l.svcCtx.DocumentTypeModel.FindOneByClientIdCode(l.ctx, clientId, req.Code)
	if err != nil {
		return &types.Response{
			Code:    response.RecordNotExistCode,
			Message: "文档类型不存在",
		}, nil
	}

	// Check if there are raw documents associated with this document type
	count, err := l.svcCtx.RawDocumentsModel.CountByDocumentCode(l.ctx, clientId, req.Code, "", "", "")
	if err != nil {
		l.Logger.Errorf("检查原始文档失败: %v", err)
		return &types.Response{
			Code:    response.ServerErrorCode,
			Message: "检查原始文档失败",
		}, nil
	}

	if count > 0 {
		return &types.Response{
			Code:    response.ConflictCode,
			Message: "该知识库下存在原始文档，无法删除",
		}, nil
	}

	// Delete document type
	err = l.svcCtx.DocumentTypeModel.DeleteByClientIdCode(l.ctx, clientId, req.Code)
	if err != nil {
		l.Logger.Errorf("删除文档类型失败: %v", err)
		return &types.Response{
			Code:    response.ServerErrorCode,
			Message: "删除文档类型失败",
			Info:    err.Error(),
		}, nil
	}

	resp = &types.Response{
		Code:    response.SuccessCode,
		Message: "删除成功",
	}
	return
}

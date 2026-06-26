package knowdata

import (
	"context"
	"strings"
	"time"

	"knowsource/api/internal/svc"
	"knowsource/api/internal/types"
	"knowsource/common/response"
	"knowsource/model"

	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateDocumentsTypeLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 更新文档类型
func NewUpdateDocumentsTypeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateDocumentsTypeLogic {
	return &UpdateDocumentsTypeLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UpdateDocumentsTypeLogic) UpdateDocumentsType(req *types.DocumentsType) (resp *types.Response, err error) {
	clientId, _ := l.ctx.Value("clientId").(string)
	clientId = strings.TrimSpace(clientId)
	if clientId == "" {
		return &types.Response{Code: response.UnauthorizedCode, Message: "clientId不能为空，请重新登录"}, nil
	}

	// Check if document type exists
	existingDocType, err := l.svcCtx.DocumentTypeModel.FindOneByClientIdCode(l.ctx, clientId, req.Code)
	if err != nil {
		return &types.Response{
			Code:    response.RecordNotExistCode,
			Message: "文档类型不存在",
		}, nil
	}

	// Update document type
	// 前端会明确传递 IsDisabled 值（0 或 1），直接使用请求的值
	data := &model.DocumentType{
		ClientId:    clientId,
		Code:        req.Code,
		Name:        req.Name,
		IsDisabled:  req.IsDisabled, // 前端会明确传递当前值
		Description: req.Description,
		CreatedAt:   existingDocType.CreatedAt,
		UpdatedAt:   time.Now(),
	}

	err = l.svcCtx.DocumentTypeModel.Update(l.ctx, data)
	if err != nil {
		l.Logger.Errorf("更新文档类型失败: %v", err)
		return &types.Response{
			Code:    response.ServerErrorCode,
			Message: "更新文档类型失败",
		}, nil
	}

	resp = &types.Response{
		Code:    response.SuccessCode,
		Message: "更新成功",
	}
	return
}

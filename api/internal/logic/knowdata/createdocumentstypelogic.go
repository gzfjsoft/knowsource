package knowdata

import (
	"context"
	"strings"
	"time"

	"knowsource/api/internal/svc"
	"knowsource/api/internal/types"
	"knowsource/api/internal/utils"
	"knowsource/common/response"
	"knowsource/model"

	"github.com/zeromicro/go-zero/core/logx"
)

type CreateDocumentsTypeLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 创建文档类型
func NewCreateDocumentsTypeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateDocumentsTypeLogic {
	return &CreateDocumentsTypeLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CreateDocumentsTypeLogic) CreateDocumentsType(req *types.DocumentsType) (resp *types.Response, err error) {
	clientId, _ := l.ctx.Value("clientId").(string)
	clientId = strings.TrimSpace(clientId)
	if clientId == "" {
		return &types.Response{
			Code:    response.UnauthorizedCode,
			Message: "clientId不能为空，请重新登录",
		}, nil
	}

	// 如果 Code 为空串，将 Name 转为拼音作为 Code
	code := req.Code
	if strings.TrimSpace(code) == "" {
		code = utils.ConvertToPinyinCode(req.Name)
		if code == "" {
			return &types.Response{
				Code:    response.ParameterErrorCode,
				Message: "Code 和 Name 不能同时为空",
			}, nil
		}
	}

	// Check if document type already exists
	_, err = l.svcCtx.DocumentTypeModel.FindOneByClientIdCode(l.ctx, clientId, code)
	if err == nil {
		return &types.Response{
			Code:    response.ConflictCode,
			Message: "文档类型已存在",
		}, nil
	}

	// Create new document type
	data := &model.DocumentType{
		ClientId:    clientId,
		Code:        code,
		Name:        req.Name,
		IsDisabled:  0, // 默认状态为正常（未禁止）
		Description: req.Description,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	_, err = l.svcCtx.DocumentTypeModel.Insert(l.ctx, data)
	if err != nil {
		l.Logger.Errorf("创建文档类型失败: %v", err)
		return &types.Response{
			Code:    response.ServerErrorCode,
			Message: "创建文档类型失败",
		}, nil
	}

	resp = &types.Response{
		Code:    response.SuccessCode,
		Message: "创建成功",
	}
	return
}

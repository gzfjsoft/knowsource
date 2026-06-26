package knowsource

import (
	"context"
	"strings"

	"knowsource/api/internal/svc"
	"knowsource/api/internal/types"
	"knowsource/common/response"
	"knowsource/model"

	"github.com/zeromicro/go-zero/core/logx"
)

type CreateEmpDocumentTypeLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 创建员工知识库权限
func NewCreateEmpDocumentTypeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateEmpDocumentTypeLogic {
	return &CreateEmpDocumentTypeLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CreateEmpDocumentTypeLogic) CreateEmpDocumentType(req *types.KnowsourceEmpDocumentTypeRequest) (resp *types.Response, err error) {
	// 验证参数
	if req.EmpCode == "" || req.DocumentTypeCode == "" {
		return &types.Response{
			Code:    response.ParameterErrorCode,
			Message: "员工编码和文档类型编码不能为空",
		}, nil
	}

	clientId, _ := l.ctx.Value("clientId").(string)
	if strings.TrimSpace(clientId) == "" {
		return &types.Response{
			Code:    response.UnauthorizedCode,
			Message: "clientId不能为空，请重新登录",
		}, nil
	}

	// 检查员工是否存在
	_, err = l.svcCtx.FrEmpModel.FindOneByClientIdFempCode(l.ctx, clientId, req.EmpCode)
	if err != nil {
		if err == model.ErrNotFound {
			return &types.Response{
				Code:    response.UserNotExistCode,
				Message: "员工不存在",
			}, nil
		}
		return &types.Response{
			Code:    response.ServerErrorCode,
			Message: err.Error(),
		}, nil
	}

	// 检查文档类型是否存在
	_, err = l.svcCtx.DocumentTypeModel.FindOneByClientIdCode(l.ctx, clientId, req.DocumentTypeCode)
	if err != nil {
		if err == model.ErrNotFound {
			return &types.Response{
				Code:    response.NotFoundCode,
				Message: "文档类型不存在",
			}, nil
		}
		return &types.Response{
			Code:    response.ServerErrorCode,
			Message: err.Error(),
		}, nil
	}

	// 检查是否已存在绑定
	_, err = l.svcCtx.EmpDocumentTypeModel.FindOneByClientIdEmpCodeDocumentTypeCode(l.ctx, clientId, req.EmpCode, req.DocumentTypeCode)
	if err == nil {
		return &types.Response{
			Code:    response.ConflictCode,
			Message: "该员工已绑定此文档类型",
		}, nil
	}
	if err != model.ErrNotFound {
		return &types.Response{
			Code:    response.ServerErrorCode,
			Message: err.Error(),
		}, nil
	}

	// 创建绑定
	empDocType := &model.EmpDocumentType{
		ClientId:         clientId,
		EmpCode:          req.EmpCode,
		DocumentTypeCode: req.DocumentTypeCode,
	}
	_, err = l.svcCtx.EmpDocumentTypeModel.Insert(l.ctx, empDocType)
	if err != nil {
		return &types.Response{
			Code:    response.ServerErrorCode,
			Message: err.Error(),
		}, nil
	}

	return &types.Response{
		Code:    response.SuccessCode,
		Message: "创建成功",
	}, nil
}

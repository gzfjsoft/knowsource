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

type CreateDeptDocumentTypeLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 创建部门文档类型绑定
func NewCreateDeptDocumentTypeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateDeptDocumentTypeLogic {
	return &CreateDeptDocumentTypeLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CreateDeptDocumentTypeLogic) CreateDeptDocumentType(req *types.KnowsourceDeptDocumentTypeRequest) (resp *types.Response, err error) {
	// 验证参数
	if req.DeptCode == "" || req.DocumentTypeCode == "" {
		return &types.Response{
			Code:    response.ParameterErrorCode,
			Message: "部门编码和文档类型编码不能为空",
		}, nil
	}

	clientId, _ := l.ctx.Value("clientId").(string)
	if strings.TrimSpace(clientId) == "" {
		return &types.Response{
			Code:    response.UnauthorizedCode,
			Message: "clientId不能为空，请重新登录",
		}, nil
	}

	// 检查部门是否存在
	_, err = l.svcCtx.FrDeptModel.FindOneByClientIdDeptCode(l.ctx, clientId, req.DeptCode)
	if err != nil {
		if err == model.ErrNotFound {
			return &types.Response{
				Code:    response.NotFoundCode,
				Message: "部门不存在",
			}, nil
		}
		return &types.Response{
			Code:    response.ServerErrorCode,
			Info:    err.Error(),
			Message: "部门查询错误",
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
	_, err = l.svcCtx.DeptDocumentTypeModel.FindOneByClientIdDeptCodeDocumentTypeCode(l.ctx, clientId, req.DeptCode, req.DocumentTypeCode)
	if err == nil {
		return &types.Response{
			Code:    response.ConflictCode,
			Message: "该部门已绑定此文档类型",
		}, nil
	}
	if err != model.ErrNotFound {
		return &types.Response{
			Code:    response.ServerErrorCode,
			Message: err.Error(),
		}, nil
	}

	// 创建绑定
	deptDocType := &model.DeptDocumentType{
		ClientId:         clientId,
		DeptCode:         req.DeptCode,
		DocumentTypeCode: req.DocumentTypeCode,
	}
	_, err = l.svcCtx.DeptDocumentTypeModel.Insert(l.ctx, deptDocType)
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

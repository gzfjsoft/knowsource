package knowsource

import (
	"context"

	"knowsource/api/internal/svc"
	"knowsource/api/internal/types"
	"knowsource/common/response"
	"knowsource/model"

	"github.com/zeromicro/go-zero/core/logx"
)

type DeleteDeptDocumentTypeLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 删除部门文档类型绑定
func NewDeleteDeptDocumentTypeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteDeptDocumentTypeLogic {
	return &DeleteDeptDocumentTypeLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *DeleteDeptDocumentTypeLogic) DeleteDeptDocumentType(req *types.KnowsourceDeptDocumentTypeDeleteRequest) (resp *types.Response, err error) {
	// 验证参数
	if req.Id <= 0 {
		return &types.Response{
			Code:    response.ParameterErrorCode,
			Message: "ID不能为空",
		}, nil
	}

	// 检查绑定是否存在
	_, err = l.svcCtx.DeptDocumentTypeModel.FindOne(l.ctx, req.Id)
	if err != nil {
		if err == model.ErrNotFound {
			return &types.Response{
				Code:    response.NotFoundCode,
				Message: "绑定记录不存在",
			}, nil
		}
		return &types.Response{
			Code:    response.ServerErrorCode,
			Message: err.Error(),
		}, nil
	}

	// 删除绑定
	err = l.svcCtx.DeptDocumentTypeModel.Delete(l.ctx, req.Id)
	if err != nil {
		return &types.Response{
			Code:    response.ServerErrorCode,
			Message: err.Error(),
		}, nil
	}

	return &types.Response{
		Code:    response.SuccessCode,
		Message: "删除成功",
	}, nil
}

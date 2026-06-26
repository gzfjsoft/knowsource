// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

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

type UpdateDeptLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 修改部门
func NewUpdateDeptLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateDeptLogic {
	return &UpdateDeptLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UpdateDeptLogic) UpdateDept(req *types.KnowsourceDeptUpdateRequest) (resp *types.Response, err error) {
	clientId, _ := l.ctx.Value("clientId").(string)
	clientId = strings.TrimSpace(clientId)
	if clientId == "" {
		return &types.Response{Code: response.UnauthorizedCode, Message: "clientId不能为空，请重新登录"}, nil
	}
	if req == nil {
		return &types.Response{Code: response.InvalidRequestParamCode, Message: "参数不能为空"}, nil
	}

	req.DeptCode = strings.TrimSpace(req.DeptCode)
	req.DeptName = strings.TrimSpace(req.DeptName)
	req.ParentCode = strings.TrimSpace(req.ParentCode)
	req.GSDM = strings.TrimSpace(req.GSDM)
	req.EndMark = strings.TrimSpace(req.EndMark)
	req.Kind = strings.TrimSpace(req.Kind)
	req.B0110 = strings.TrimSpace(req.B0110)

	if req.DeptCode == "" || req.DeptName == "" {
		return &types.Response{Code: response.InvalidRequestParamCode, Message: "deptCode/deptName不能为空"}, nil
	}

	row, fErr := l.svcCtx.FrDeptModel.FindOneByClientIdDeptCode(l.ctx, clientId, req.DeptCode)
	if fErr != nil {
		if fErr == model.ErrNotFound {
			return &types.Response{Code: response.NotFoundCode, Message: "部门不存在"}, nil
		}
		return &types.Response{Code: response.ServerErrorCode, Message: "Database error", Info: fErr.Error()}, nil
	}

	if req.ParentCode == "" {
		req.ParentCode = row.ParentCode
	}
	if req.EndMark == "" {
		req.EndMark = row.EndMark
	}

	row.Gsdm = req.GSDM
	row.DeptName = req.DeptName
	row.ParentCode = req.ParentCode
	row.EndMark = req.EndMark
	row.GRADE = req.Grade
	row.Kind = req.Kind
	row.B0110 = req.B0110

	if uErr := l.svcCtx.FrDeptModel.Update(l.ctx, row); uErr != nil {
		return &types.Response{Code: response.ServerErrorCode, Message: "Database error", Info: uErr.Error()}, nil
	}

	return &types.Response{Code: response.SuccessCode, Message: "success"}, nil

}

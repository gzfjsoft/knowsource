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

type CreateDeptLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 新增部门
func NewCreateDeptLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateDeptLogic {
	return &CreateDeptLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CreateDeptLogic) CreateDept(req *types.KnowsourceDeptCreateRequest) (resp *types.Response, err error) {
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

	if _, fErr := l.svcCtx.FrDeptModel.FindOneByClientIdDeptCode(l.ctx, clientId, req.DeptCode); fErr == nil {
		return &types.Response{Code: response.ConflictCode, Message: "部门已存在"}, nil
	} else if fErr != model.ErrNotFound {
		return &types.Response{Code: response.ServerErrorCode, Message: "Database error", Info: fErr.Error()}, nil
	}

	if req.ParentCode == "" {
		req.ParentCode = "0"
	}
	if req.EndMark == "" {
		req.EndMark = "0"
	}

	_, iErr := l.svcCtx.FrDeptModel.Insert(l.ctx, &model.FrDept{
		ClientId:   clientId,
		Gsdm:       req.GSDM,
		DeptCode:   req.DeptCode,
		DeptName:   req.DeptName,
		ParentCode: req.ParentCode,
		EndMark:    req.EndMark,
		GRADE:      req.Grade,
		Kind:       req.Kind,
		B0110:      req.B0110,
	})
	if iErr != nil {
		return &types.Response{Code: response.ServerErrorCode, Message: "Database error", Info: iErr.Error()}, nil
	}

	return &types.Response{Code: response.SuccessCode, Message: "success"}, nil

}

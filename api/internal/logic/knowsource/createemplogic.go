// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package knowsource

import (
	"context"
	"database/sql"
	"strings"

	"knowsource/api/internal/svc"
	"knowsource/api/internal/types"
	"knowsource/common/response"
	"knowsource/model"

	"github.com/zeromicro/go-zero/core/logx"
)

type CreateEmpLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 新增员工
func NewCreateEmpLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateEmpLogic {
	return &CreateEmpLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CreateEmpLogic) CreateEmp(req *types.KnowsourceEmpCreateRequest) (resp *types.Response, err error) {
	clientId, _ := l.ctx.Value("clientId").(string)
	clientId = strings.TrimSpace(clientId)
	if clientId == "" {
		return &types.Response{Code: response.UnauthorizedCode, Message: "clientId不能为空，请重新登录"}, nil
	}
	if req == nil {
		return &types.Response{Code: response.InvalidRequestParamCode, Message: "参数不能为空"}, nil
	}

	req.EmpCode = strings.TrimSpace(req.EmpCode)
	req.EmpName = strings.TrimSpace(req.EmpName)
	req.DeptCode = strings.TrimSpace(req.DeptCode)
	req.Position = strings.TrimSpace(req.Position)
	req.Branch = strings.TrimSpace(req.Branch)

	if req.EmpCode == "" || req.EmpName == "" || req.DeptCode == "" {
		return &types.Response{Code: response.InvalidRequestParamCode, Message: "empCode/empName/deptCode不能为空"}, nil
	}

	// 校验部门存在
	if _, dErr := l.svcCtx.FrDeptModel.FindOneByClientIdDeptCode(l.ctx, clientId, req.DeptCode); dErr != nil {
		if dErr == model.ErrNotFound {
			return &types.Response{Code: response.InvalidRequestParamCode, Message: "deptCode不存在"}, nil
		}
		return &types.Response{Code: response.ServerErrorCode, Message: "Database error", Info: dErr.Error()}, nil
	}

	if _, fErr := l.svcCtx.FrEmpModel.FindOneByClientIdFempCode(l.ctx, clientId, req.EmpCode); fErr == nil {
		return &types.Response{Code: response.ConflictCode, Message: "员工已存在"}, nil
	} else if fErr != model.ErrNotFound {
		return &types.Response{Code: response.ServerErrorCode, Message: "Database error", Info: fErr.Error()}, nil
	}

	status := req.Status
	if status != 0 && status != 1 {
		status = 0
	}

	em := strings.TrimSpace(strings.ToLower(req.Email))
	ph := strings.TrimSpace(req.Mobile)
	if em != "" {
		if o, e := l.svcCtx.FrEmpModel.FindOneByClientIdEmail(l.ctx, clientId, em); e == nil && o != nil {
			return &types.Response{Code: response.ConflictCode, Message: "该邮箱已被其他员工使用"}, nil
		} else if e != nil && e != model.ErrNotFound {
			return &types.Response{Code: response.ServerErrorCode, Message: "Database error", Info: e.Error()}, nil
		}
	}
	if ph != "" {
		if o, e := l.svcCtx.FrEmpModel.FindOneByClientIdMobile(l.ctx, clientId, ph); e == nil && o != nil {
			return &types.Response{Code: response.ConflictCode, Message: "该手机号已被其他员工使用"}, nil
		} else if e != nil && e != model.ErrNotFound {
			return &types.Response{Code: response.ServerErrorCode, Message: "Database error", Info: e.Error()}, nil
		}
	}
	_, iErr := l.svcCtx.FrEmpModel.Insert(l.ctx, &model.FrEmp{
		ClientId:  clientId,
		Frylb:     "",
		Status:    status,
		FempName:  req.EmpName,
		DeptCode:  req.DeptCode,
		FdeptId:   sql.NullInt64{Valid: false},
		FempCode:  req.EmpCode,
		Fposition: req.Position,
		Fbranch:   req.Branch,
		Mobile:    ph,
		Email:     em,
	})
	if iErr != nil {
		return &types.Response{Code: response.ServerErrorCode, Message: "Database error", Info: iErr.Error()}, nil
	}

	return &types.Response{Code: response.SuccessCode, Message: "success"}, nil

}

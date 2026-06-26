// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package knowsource

import (
	"context"
	"strings"

	"knowsource/api/internal/svc"
	"knowsource/api/internal/types"
	"knowsource/common/response"
	"knowsource/consts"
	"knowsource/model"

	"github.com/zeromicro/go-zero/core/logx"
)

type DeleteEmpLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 删除员工
func NewDeleteEmpLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteEmpLogic {
	return &DeleteEmpLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *DeleteEmpLogic) DeleteEmp(req *types.KnowsourceEmpDeleteRequest) (resp *types.Response, err error) {
	clientId, _ := l.ctx.Value("clientId").(string)
	clientId = strings.TrimSpace(clientId)
	if clientId == "" {
		return &types.Response{Code: response.UnauthorizedCode, Message: "clientId不能为空，请重新登录"}, nil
	}
	if req == nil {
		return &types.Response{Code: response.InvalidRequestParamCode, Message: "参数不能为空"}, nil
	}

	req.EmpCode = strings.TrimSpace(req.EmpCode)
	if req.EmpCode == "" {
		return &types.Response{Code: response.InvalidRequestParamCode, Message: "empCode不能为空"}, nil
	}

	row, fErr := l.svcCtx.FrEmpModel.FindOneByClientIdFempCode(l.ctx, clientId, req.EmpCode)
	if fErr != nil {
		if fErr == model.ErrNotFound {
			return &types.Response{Code: response.NotFoundCode, Message: "员工不存在"}, nil
		}
		return &types.Response{Code: response.ServerErrorCode, Message: "Database error", Info: fErr.Error()}, nil
	}

	// 保护：该租户仅剩一个 superadmin 角色用户时禁止删除
	roles, rErr := l.svcCtx.FrUserRolesModel.FindAllByClientIdEmpCode(l.ctx, clientId, req.EmpCode)
	if rErr != nil {
		return &types.Response{Code: response.ServerErrorCode, Message: "Database error", Info: rErr.Error()}, nil
	}
	isSuperadmin := false
	for _, rr := range roles {
		if rr != nil && rr.Role == consts.SUPER_ADMIN {
			isSuperadmin = true
			break
		}
	}
	if isSuperadmin {
		cnt, cErr := l.svcCtx.FrUserRolesModel.CountByClientIdRole(l.ctx, clientId, consts.SUPER_ADMIN)
		if cErr != nil {
			return &types.Response{Code: response.ServerErrorCode, Message: "Database error", Info: cErr.Error()}, nil
		}
		if cnt <= 1 {
			return &types.Response{Code: response.ConflictCode, Message: "该租户仅剩一个superadmin用户，禁止删除"}, nil
		}
	}

	if dErr := l.svcCtx.FrEmpModel.Delete(l.ctx, row.Id); dErr != nil {
		return &types.Response{Code: response.ServerErrorCode, Message: "Database error", Info: dErr.Error()}, nil
	}

	return &types.Response{Code: response.SuccessCode, Message: "success"}, nil

}

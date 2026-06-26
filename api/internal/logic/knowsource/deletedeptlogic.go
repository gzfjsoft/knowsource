// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package knowsource

import (
	"context"
	"fmt"
	"strings"

	"knowsource/api/internal/svc"
	"knowsource/api/internal/types"
	"knowsource/common/response"
	"knowsource/model"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type DeleteDeptLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 删除部门
func NewDeleteDeptLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteDeptLogic {
	return &DeleteDeptLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *DeleteDeptLogic) DeleteDept(req *types.KnowsourceDeptDeleteRequest) (resp *types.Response, err error) {
	clientId, _ := l.ctx.Value("clientId").(string)
	clientId = strings.TrimSpace(clientId)
	if clientId == "" {
		return &types.Response{Code: response.UnauthorizedCode, Message: "clientId不能为空，请重新登录"}, nil
	}
	if req == nil {
		return &types.Response{Code: response.InvalidRequestParamCode, Message: "参数不能为空"}, nil
	}

	req.DeptCode = strings.TrimSpace(req.DeptCode)
	if req.DeptCode == "" {
		return &types.Response{Code: response.InvalidRequestParamCode, Message: "deptCode不能为空"}, nil
	}

	// 确认部门存在
	if _, fErr := l.svcCtx.FrDeptModel.FindOneByClientIdDeptCode(l.ctx, clientId, req.DeptCode); fErr != nil {
		if fErr == model.ErrNotFound {
			return &types.Response{Code: response.NotFoundCode, Message: "部门不存在"}, nil
		}
		return &types.Response{Code: response.ServerErrorCode, Message: "Database error", Info: fErr.Error()}, nil
	}

	// 非级联删除：若存在子部门则拒绝
	if req.Cascade != 1 {
		hasChild, hErr := l.svcCtx.FrDeptModel.HasChildren(l.ctx, clientId, req.DeptCode)
		if hErr != nil {
			return &types.Response{Code: response.ServerErrorCode, Message: "Database error", Info: hErr.Error()}, nil
		}
		if hasChild {
			return &types.Response{Code: response.ConflictCode, Message: "该部门存在子部门，不能直接删除"}, nil
		}

		if dErr := l.svcCtx.FrDeptModel.DeleteByClientIdDeptCodes(l.ctx, clientId, []string{req.DeptCode}); dErr != nil {
			return &types.Response{Code: response.ServerErrorCode, Message: "Database error", Info: dErr.Error()}, nil
		}
		return &types.Response{Code: response.SuccessCode, Message: "success"}, nil
	}

	// 级联删除：删除整棵子树（含自己）
	rows, qErr := l.svcCtx.FrDeptModel.FindAllLiteByClientId(l.ctx, clientId)
	if qErr != nil && qErr != sqlx.ErrNotFound {
		return &types.Response{Code: response.ServerErrorCode, Message: "Database error", Info: qErr.Error()}, nil
	}

	childrenMap := make(map[string][]string, len(rows))
	exists := make(map[string]bool, len(rows))
	for _, r := range rows {
		if r == nil {
			continue
		}
		exists[r.DeptCode] = true
		childrenMap[r.ParentCode] = append(childrenMap[r.ParentCode], r.DeptCode)
	}
	if !exists[req.DeptCode] {
		return &types.Response{Code: response.NotFoundCode, Message: "部门不存在"}, nil
	}

	var toDelete []string
	queue := []string{req.DeptCode}
	seen := map[string]bool{req.DeptCode: true}
	for len(queue) > 0 {
		cur := queue[0]
		queue = queue[1:]
		toDelete = append(toDelete, cur)
		for _, child := range childrenMap[cur] {
			if !seen[child] {
				seen[child] = true
				queue = append(queue, child)
			}
		}
	}

	// 事务内删除（叶子优先不是必须，但可读性更好）
	for i, j := 0, len(toDelete)-1; i < j; i, j = i+1, j-1 {
		toDelete[i], toDelete[j] = toDelete[j], toDelete[i]
	}

	err = l.svcCtx.Mysql.TransactCtx(l.ctx, func(ctx context.Context, s sqlx.Session) error {
		deptModel := l.svcCtx.FrDeptModel.WithSession(s)
		if dErr := deptModel.DeleteByClientIdDeptCodes(ctx, clientId, toDelete); dErr != nil {
			return fmt.Errorf("DeleteDept cascade delete failed: %w", dErr)
		}
		return nil
	})
	if err != nil {
		return &types.Response{Code: response.ServerErrorCode, Message: "Database error", Info: err.Error()}, nil
	}

	return &types.Response{Code: response.SuccessCode, Message: "success"}, nil

}

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

type MoveDeptLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 移动部门（拖拽）
func NewMoveDeptLogic(ctx context.Context, svcCtx *svc.ServiceContext) *MoveDeptLogic {
	return &MoveDeptLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *MoveDeptLogic) MoveDept(req *types.KnowsourceDeptMoveRequest) (resp *types.Response, err error) {
	clientId, _ := l.ctx.Value("clientId").(string)
	clientId = strings.TrimSpace(clientId)
	if clientId == "" {
		return &types.Response{Code: response.UnauthorizedCode, Message: "clientId不能为空，请重新登录"}, nil
	}
	if req == nil {
		return &types.Response{Code: response.InvalidRequestParamCode, Message: "参数不能为空"}, nil
	}

	req.DeptCode = strings.TrimSpace(req.DeptCode)
	req.NewParentCode = strings.TrimSpace(req.NewParentCode)
	if req.DeptCode == "" {
		return &types.Response{Code: response.InvalidRequestParamCode, Message: "deptCode不能为空"}, nil
	}
	if req.NewParentCode == req.DeptCode {
		return &types.Response{Code: response.InvalidRequestParamCode, Message: "不能移动到自身下面"}, nil
	}

	// 校验部门存在
	if _, fErr := l.svcCtx.FrDeptModel.FindOneByClientIdDeptCode(l.ctx, clientId, req.DeptCode); fErr != nil {
		if fErr == model.ErrNotFound {
			return &types.Response{Code: response.NotFoundCode, Message: "部门不存在"}, nil
		}
		return &types.Response{Code: response.ServerErrorCode, Message: "Database error", Info: fErr.Error()}, nil
	}

	var parentGrade int64 = -1
	if req.NewParentCode != "" {
		parent, pErr := l.svcCtx.FrDeptModel.FindOneByClientIdDeptCode(l.ctx, clientId, req.NewParentCode)
		if pErr != nil {
			if pErr == model.ErrNotFound {
				return &types.Response{Code: response.InvalidRequestParamCode, Message: "newParentCode不存在"}, nil
			}
			return &types.Response{Code: response.ServerErrorCode, Message: "Database error", Info: pErr.Error()}, nil
		}
		parentGrade = parent.GRADE
	}

	rows, qErr := l.svcCtx.FrDeptModel.FindAllLiteByClientId(l.ctx, clientId)
	if qErr != nil && qErr != sqlx.ErrNotFound {
		return &types.Response{Code: response.ServerErrorCode, Message: "Database error", Info: qErr.Error()}, nil
	}

	parentMap := make(map[string]string, len(rows))
	childrenMap := make(map[string][]string, len(rows))
	for _, r := range rows {
		if r == nil {
			continue
		}
		parentMap[r.DeptCode] = r.ParentCode
		childrenMap[r.ParentCode] = append(childrenMap[r.ParentCode], r.DeptCode)
	}

	// 禁止把节点移动到自己的子树里
	subtree := map[string]bool{req.DeptCode: true}
	queue := []string{req.DeptCode}
	for len(queue) > 0 {
		cur := queue[0]
		queue = queue[1:]
		for _, child := range childrenMap[cur] {
			if !subtree[child] {
				subtree[child] = true
				queue = append(queue, child)
			}
		}
	}
	if req.NewParentCode != "" && subtree[req.NewParentCode] {
		return &types.Response{Code: response.InvalidRequestParamCode, Message: "不能移动到自己的子部门下面"}, nil
	}

	// 计算新的 grade：newRootGrade = parentGrade + 1（根则 0）
	newRootGrade := int64(0)
	if parentGrade >= 0 {
		newRootGrade = parentGrade + 1
	}

	// 计算子树内每个节点相对 root 的深度
	depth := map[string]int64{req.DeptCode: 0}
	queue = []string{req.DeptCode}
	for len(queue) > 0 {
		cur := queue[0]
		queue = queue[1:]
		for _, child := range childrenMap[cur] {
			if !subtree[child] {
				continue
			}
			if _, ok := depth[child]; ok {
				continue
			}
			depth[child] = depth[cur] + 1
			queue = append(queue, child)
		}
	}

	// 事务更新：root 更新 parent_code + grade；子节点仅更新 grade
	err = l.svcCtx.Mysql.TransactCtx(l.ctx, func(ctx context.Context, s sqlx.Session) error {
		deptModel := l.svcCtx.FrDeptModel.WithSession(s)
		if uErr := deptModel.UpdateParentAndGradeByClientIdDeptCode(ctx, clientId, req.DeptCode, req.NewParentCode, newRootGrade); uErr != nil {
			return fmt.Errorf("update root dept failed: %w", uErr)
		}
		for code, d := range depth {
			if code == req.DeptCode {
				continue
			}
			if uErr := deptModel.UpdateGradeByClientIdDeptCode(ctx, clientId, code, newRootGrade+d); uErr != nil {
				return fmt.Errorf("update dept grade failed, dept=%s: %w", code, uErr)
			}
		}
		return nil
	})
	if err != nil {
		return &types.Response{Code: response.ServerErrorCode, Message: "Database error", Info: err.Error()}, nil
	}

	_ = parentMap // 保留以便后续扩展（例如校验更多关系），避免静态检查误报
	return &types.Response{Code: response.SuccessCode, Message: "success"}, nil
}

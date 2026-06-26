// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package knowsource

import (
	"context"
	"strings"

	"knowsource/api/internal/svc"
	"knowsource/api/internal/types"
	"knowsource/common/response"
	"knowsource/consts"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type AdminClientDeleteLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// admin 删除 client
func NewAdminClientDeleteLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AdminClientDeleteLogic {
	return &AdminClientDeleteLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *AdminClientDeleteLogic) AdminClientDelete(req *types.AdminClientDeleteRequest) (resp *types.Response, err error) {
	// 1. 检查用户角色
	rolesStr, _ := l.ctx.Value("roles").(string)
	if !strings.Contains(rolesStr, consts.SUPER_ADMIN) {
		return &types.Response{
			Code:    response.UnauthorizedCode,
			Message: "没有权限",
		}, nil
	}

	// 2. 检查请求参数
	if strings.TrimSpace(req.ClientId) == "" {
		return &types.Response{
			Code:    response.ParameterErrorCode,
			Message: "clientId不能为空",
		}, nil
	}

	clientId := strings.TrimSpace(req.ClientId)

	// 3. 检查是否为当前登录用户的clientId
	currentClientId, _ := l.ctx.Value("clientId").(string)
	if currentClientId != clientId {
		return &types.Response{
			Code:    response.ForbiddenCode,
			Message: "只能删除当前登录租户的clientId",
		}, nil
	}

	// 4. 保护机制：禁止删除demo和admin租户
	if clientId == "demo" || clientId == consts.ONLY_ADMIN {
		return &types.Response{
			Code:    response.ForbiddenCode,
			Message: "禁止删除系统保护的租户",
		}, nil
	}

	// 5. 在事务中删除所有与clientId相关的数据
	err = l.svcCtx.Mysql.TransactCtx(l.ctx, func(ctx context.Context, s sqlx.Session) error {
		conn := sqlx.NewSqlConnFromSession(s)

		// 1. 删除员工知识库权限
		if _, e := conn.ExecCtx(ctx, "DELETE FROM emp_document_type WHERE client_id = ?", clientId); e != nil {
			return e
		}

		// 2. 删除部门文档类型绑定
		if _, e := conn.ExecCtx(ctx, "DELETE FROM dept_document_type WHERE client_id = ?", clientId); e != nil {
			return e
		}

		// 3. 删除用户角色关联
		if _, e := conn.ExecCtx(ctx, "DELETE FROM fr_user_roles WHERE client_id = ?", clientId); e != nil {
			return e
		}

		// 4. 删除角色权限关联
		if _, e := conn.ExecCtx(ctx, "DELETE FROM fr_roles_permissions WHERE client_id = ?", clientId); e != nil {
			return e
		}

		// 5. 删除角色
		if _, e := conn.ExecCtx(ctx, "DELETE FROM fr_roles WHERE client_id = ?", clientId); e != nil {
			return e
		}

		// 7. 删除员工密码
		if _, e := conn.ExecCtx(ctx, "DELETE FROM emp_password WHERE client_id = ?", clientId); e != nil {
			return e
		}

		// 8. 删除员工
		if _, e := conn.ExecCtx(ctx, "DELETE FROM fr_emp WHERE client_id = ?", clientId); e != nil {
			return e
		}

		// 9. 删除部门
		if _, e := conn.ExecCtx(ctx, "DELETE FROM fr_dept WHERE client_id = ?", clientId); e != nil {
			return e
		}

		// 10. 删除AI配置
		if _, e := conn.ExecCtx(ctx, "DELETE FROM ai_config WHERE client_id = ?", clientId); e != nil {
			return e
		}

		// 11. 删除Dify配置
		if _, e := conn.ExecCtx(ctx, "DELETE FROM dify_options WHERE client_id = ?", clientId); e != nil {
			return e
		}

		// 12. 最后删除client本身
		if _, e := conn.ExecCtx(ctx, "DELETE FROM client WHERE client_id = ?", clientId); e != nil {
			return e
		}

		return nil
	})

	if err != nil {
		logx.Errorf("删除client %s 失败: %v", clientId, err)
		return &types.Response{
			Code:    response.ServerErrorCode,
			Message: "删除失败",
			Info:    err.Error(),
		}, nil
	}

	return &types.Response{
		Code:    response.SuccessCode,
		Message: "删除成功",
	}, nil
}

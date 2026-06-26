package logic

import (
	"context"
	"knowsource/api/internal/svc"
	"knowsource/common/response"
	"knowsource/consts"
	"knowsource/model"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type DeleteRoleLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewDeleteRoleLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteRoleLogic {
	return &DeleteRoleLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *DeleteRoleLogic) DeleteRole(id uint64) (resp response.Response) {
	sysrole, _ := l.ctx.Value("role").(string)
	if sysrole != consts.ONLY_ADMIN && sysrole != consts.SUPER_ADMIN {
		var res response.Response
		res.Code = response.UnauthorizedCode
		res.Message = "only admin can access"
		return res

	}
	// Check if role exists
	role, err := l.svcCtx.RolesModel.FindOne(l.ctx, id)
	if err != nil {
		if err == model.ErrNotFound {
			return response.Fail(response.InvalidRequestParamCode, "Role not found")
		}
		return response.Error(err.Error())
	}

	// Check if there are users associated with this role
	userCount, err := l.svcCtx.UserRolesModel.CountByRoleId(l.ctx, role.RoleId)
	if err != nil {
		l.Logger.Errorf("查询角色关联用户数失败: %v", err)
		return response.Error("查询角色关联用户数失败")
	}
	if userCount > 0 {
		return response.Fail(response.InvalidRequestParamCode, "该角色还有关联的用户，无法删除。请先解除该角色与用户的关联关系")
	}

	err = sqlx.NewMysql(l.svcCtx.Config.MySQL.DataSource).Transact(func(session sqlx.Session) error {
		// Delete role permissions first
		rolePermissionsModel := l.svcCtx.RolePermissionsModel.WithSession(session)
		err := rolePermissionsModel.DeleteByRoleId(l.ctx, role.RoleId)
		if err != nil {
			return err
		}

		// Delete user roles
		userRolesModel := l.svcCtx.UserRolesModel.WithSession(session)
		err = userRolesModel.DeleteByRoleId(l.ctx, role.RoleId)
		if err != nil {
			return err
		}

		// Delete role
		rolesModel := l.svcCtx.RolesModel.WithSession(session)
		return rolesModel.Delete(l.ctx, role.RoleId)
	})

	if err != nil {
		return response.Error(err.Error())
	}

	return response.OK(nil)
}

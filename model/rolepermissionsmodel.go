package model

import (
	"context"
	"fmt"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ RolePermissionsModel = (*customRolePermissionsModel)(nil)

type (
	// RolePermissionsModel is an interface to be customized, add more methods here,
	// and implement the added methods in customRolePermissionsModel.
	RolePermissionsModel interface {
		rolePermissionsModel
		WithSession(session sqlx.Session) RolePermissionsModel
		DeleteByRoleId(ctx context.Context, roleId uint64) error
		DeleteByPermissionId(ctx context.Context, permissionId uint64) error
		FindByRoleId(ctx context.Context, roleId uint64) ([]*RolePermissions, error)
		FixMissingRolePermissions(ctx context.Context) error
	}

	customRolePermissionsModel struct {
		*defaultRolePermissionsModel
	}
)

// NewRolePermissionsModel returns a model for the database table.
func NewRolePermissionsModel(conn sqlx.SqlConn) RolePermissionsModel {
	return &customRolePermissionsModel{
		defaultRolePermissionsModel: newRolePermissionsModel(conn),
	}
}

func (m *customRolePermissionsModel) WithSession(session sqlx.Session) RolePermissionsModel {
	return NewRolePermissionsModel(sqlx.NewSqlConnFromSession(session))
}

func (m *customRolePermissionsModel) DeleteByRoleId(ctx context.Context, roleId uint64) error {
	query := fmt.Sprintf("delete from %s where role_id = ?", m.table)
	_, err := m.conn.ExecCtx(ctx, query, roleId)
	return err
}

func (m *customRolePermissionsModel) DeleteByPermissionId(ctx context.Context, permissionId uint64) error {
	query := fmt.Sprintf("delete from %s where permission_id = ?", m.table)
	_, err := m.conn.ExecCtx(ctx, query, permissionId)
	return err
}

func (m *customRolePermissionsModel) FindByRoleId(ctx context.Context, roleId uint64) ([]*RolePermissions, error) {
	var resp []*RolePermissions
	err := m.conn.QueryRowsCtx(ctx, &resp, "SELECT * FROM role_permissions WHERE role_id = ?", roleId)
	return resp, err
}

func (m *customRolePermissionsModel) FixMissingRolePermissions(ctx context.Context) error {
	sql := `insert into 
  role_permissions (
     
    role_id, 
    permission_name, 
    granted_at, 
    granted_by
  ) select 1 as role_id,permission_name,now(),1 as granted_by  from permissions where permission_name not in (select permission_name from role_permissions where role_id=1)`
	_, err := m.conn.ExecCtx(ctx, sql)
	return err
}

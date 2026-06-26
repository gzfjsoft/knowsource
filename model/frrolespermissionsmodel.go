package model

import (
	"context"
	"fmt"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ FrRolesPermissionsModel = (*customFrRolesPermissionsModel)(nil)

type (
	// FrRolesPermissionsModel is an interface to be customized, add more methods here,
	// and implement the added methods in customFrRolesPermissionsModel.
	FrRolesPermissionsModel interface {
		frRolesPermissionsModel
		withSession(session sqlx.Session) FrRolesPermissionsModel
		FindAllByClientIdEmpCode(ctx context.Context, clientId, empCode string) ([]*FrPermissionsOnly, error)
	}

	customFrRolesPermissionsModel struct {
		*defaultFrRolesPermissionsModel
	}
)

// NewFrRolesPermissionsModel returns a model for the database table.
func NewFrRolesPermissionsModel(conn sqlx.SqlConn) FrRolesPermissionsModel {
	return &customFrRolesPermissionsModel{
		defaultFrRolesPermissionsModel: newFrRolesPermissionsModel(conn),
	}
}

func (m *customFrRolesPermissionsModel) withSession(session sqlx.Session) FrRolesPermissionsModel {
	return NewFrRolesPermissionsModel(sqlx.NewSqlConnFromSession(session))
}

type FrPermissionsOnly struct {
	Permission string `db:"permission"` // 权限
}

func (m *customFrRolesPermissionsModel) FindAllByClientIdEmpCode(ctx context.Context, clientId, empCode string) ([]*FrPermissionsOnly, error) {

	var resp []*FrPermissionsOnly
	query := fmt.Sprintf("select distinct permission from %s where `client_id` = ? AND `role` in (select role from fr_user_roles where client_id = ? AND emp_code = ?)", m.table)
	err := m.conn.QueryRowsCtx(ctx, &resp, query, clientId, clientId, empCode)
	switch err {
	case nil:
		return resp, nil
	case sqlx.ErrNotFound:
		return resp, nil
	default:
		return nil, err
	}
}

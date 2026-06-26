package model

import (
	"context"
	"fmt"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ FrRolesModel = (*customFrRolesModel)(nil)

type (
	FrRolesModel interface {
		frRolesModel
		withSession(session sqlx.Session) FrRolesModel
		FindOneByClientIdRole(ctx context.Context, clientId, role string) (*FrRoles, error)
		DeleteByClientIdRole(ctx context.Context, clientId, role string) error
	}

	customFrRolesModel struct {
		*defaultFrRolesModel
	}
)

func NewFrRolesModel(conn sqlx.SqlConn) FrRolesModel {
	return &customFrRolesModel{
		defaultFrRolesModel: newFrRolesModel(conn),
	}
}

func (m *customFrRolesModel) withSession(session sqlx.Session) FrRolesModel {
	return NewFrRolesModel(sqlx.NewSqlConnFromSession(session))
}

func (m *customFrRolesModel) FindOneByClientIdRole(ctx context.Context, clientId, role string) (*FrRoles, error) {
	query := fmt.Sprintf("SELECT %s FROM %s WHERE `client_id` = ? AND `role` = ? LIMIT 1", frRolesRows, m.table)
	var resp FrRoles
	err := m.conn.QueryRowCtx(ctx, &resp, query, clientId, role)
	switch err {
	case nil:
		return &resp, nil
	case sqlx.ErrNotFound:
		return nil, ErrNotFound
	default:
		return nil, err
	}
}

func (m *customFrRolesModel) DeleteByClientIdRole(ctx context.Context, clientId, role string) error {
	query := fmt.Sprintf("DELETE FROM %s WHERE `client_id` = ? AND `role` = ?", m.table)
	_, err := m.conn.ExecCtx(ctx, query, clientId, role)
	return err
}

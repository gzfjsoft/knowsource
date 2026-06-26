package model

import (
	"context"
	"fmt"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ FrUserRolesModel = (*customFrUserRolesModel)(nil)

type (
	// FrUserRolesModel is an interface to be customized, add more methods here,
	// and implement the added methods in customFrUserRolesModel.
	FrUserRolesModel interface {
		frUserRolesModel
		withSession(session sqlx.Session) FrUserRolesModel
		FindAllByClientIdEmpCode(ctx context.Context, clientId, empCode string) ([]*FrUserRoles, error)
		CountByClientIdRole(ctx context.Context, clientId, role string) (int64, error)
		CountListByClientId(ctx context.Context, clientId, empCode, role string) (int64, error)
		FindListByClientId(ctx context.Context, clientId, empCode, role string, limit, offset int64) ([]*FrUserRoles, error)
	}

	customFrUserRolesModel struct {
		*defaultFrUserRolesModel
	}
)

// NewFrUserRolesModel returns a model for the database table.
func NewFrUserRolesModel(conn sqlx.SqlConn) FrUserRolesModel {
	return &customFrUserRolesModel{
		defaultFrUserRolesModel: newFrUserRolesModel(conn),
	}
}

func (m *customFrUserRolesModel) withSession(session sqlx.Session) FrUserRolesModel {
	return NewFrUserRolesModel(sqlx.NewSqlConnFromSession(session))
}

func (m *customFrUserRolesModel) FindAllByClientIdEmpCode(ctx context.Context, clientId, empCode string) ([]*FrUserRoles, error) {
	var resp []*FrUserRoles
	query := fmt.Sprintf("select %s from %s where `client_id` = ? AND `emp_code` = ?", frUserRolesRows, m.table)
	err := m.conn.QueryRowsCtx(ctx, &resp, query, clientId, empCode)
	switch err {
	case nil:
		return resp, nil
	case sqlx.ErrNotFound:
		return resp, nil
	default:
		return nil, err
	}

}

func (m *customFrUserRolesModel) CountByClientIdRole(ctx context.Context, clientId, role string) (int64, error) {
	var count int64
	query := fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE `role` = ? AND `client_id` = ?", m.table)
	err := m.conn.QueryRowCtx(ctx, &count, query, role, clientId)
	if err != nil {
		return 0, err
	}
	return count, nil
}

// CountListByClientId 按租户统计用户角色关联条数，empCode/role 为空则不作为条件。
func (m *customFrUserRolesModel) CountListByClientId(ctx context.Context, clientId, empCode, role string) (int64, error) {
	query := fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE `client_id` = ?", m.table)
	args := []interface{}{clientId}
	if empCode != "" {
		query += " AND `emp_code` = ?"
		args = append(args, empCode)
	}
	if role != "" {
		query += " AND `role` = ?"
		args = append(args, role)
	}
	var total int64
	err := m.conn.QueryRowCtx(ctx, &total, query, args...)
	if err != nil {
		return 0, err
	}
	return total, nil
}

// FindListByClientId 按租户分页查询用户角色关联，empCode/role 为空则不作为条件。
func (m *customFrUserRolesModel) FindListByClientId(ctx context.Context, clientId, empCode, role string, limit, offset int64) ([]*FrUserRoles, error) {
	query := fmt.Sprintf("SELECT %s FROM %s WHERE `client_id` = ?", frUserRolesRows, m.table)
	args := []interface{}{clientId}
	if empCode != "" {
		query += " AND `emp_code` = ?"
		args = append(args, empCode)
	}
	if role != "" {
		query += " AND `role` = ?"
		args = append(args, role)
	}
	query += " ORDER BY `id` DESC LIMIT ? OFFSET ?"
	args = append(args, limit, offset)

	var resp []*FrUserRoles
	err := m.conn.QueryRowsCtx(ctx, &resp, query, args...)
	switch err {
	case nil:
		return resp, nil
	case sqlx.ErrNotFound:
		return []*FrUserRoles{}, nil
	default:
		return nil, err
	}
}

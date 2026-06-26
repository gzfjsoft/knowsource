package model

import (
	"context"
	"fmt"
	"strings"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ FrEmpModel = (*customFrEmpModel)(nil)

type (
	// FrEmpModel is an interface to be customized, add more methods here,
	// and implement the added methods in customFrEmpModel.
	FrEmpModel interface {
		frEmpModel
		withSession(session sqlx.Session) FrEmpModel
		FindOneByClientIdFempCode(ctx context.Context, clientId, fempCode string) (*FrEmp, error)
		FindOneByClientIdEmail(ctx context.Context, clientId, email string) (*FrEmp, error)
		FindOneByClientIdMobile(ctx context.Context, clientId, mobile string) (*FrEmp, error)
	}

	customFrEmpModel struct {
		*defaultFrEmpModel
	}
)

// NewFrEmpModel returns a model for the database table.
func NewFrEmpModel(conn sqlx.SqlConn) FrEmpModel {
	return &customFrEmpModel{
		defaultFrEmpModel: newFrEmpModel(conn),
	}
}

func (m *customFrEmpModel) withSession(session sqlx.Session) FrEmpModel {
	return NewFrEmpModel(sqlx.NewSqlConnFromSession(session))
}

func (m *customFrEmpModel) FindOneByClientIdFempCode(ctx context.Context, clientId, fempCode string) (*FrEmp, error) {
	clientId = strings.TrimSpace(clientId)
	fempCode = strings.TrimSpace(fempCode)
	if clientId == "" || fempCode == "" {
		return nil, ErrNotFound
	}
	var resp FrEmp
	q := fmt.Sprintf("SELECT %s FROM %s WHERE `client_id`=? AND `femp_code`=? LIMIT 1", frEmpRows, m.table)
	err := m.conn.QueryRowCtx(ctx, &resp, q, clientId, fempCode)
	switch err {
	case nil:
		return &resp, nil
	case sqlx.ErrNotFound:
		return nil, ErrNotFound
	default:
		return nil, err
	}
}

func (m *customFrEmpModel) FindOneByClientIdEmail(ctx context.Context, clientId, email string) (*FrEmp, error) {
	clientId = strings.TrimSpace(clientId)
	email = strings.TrimSpace(email)
	if clientId == "" || email == "" {
		return nil, ErrNotFound
	}
	var resp FrEmp
	q := fmt.Sprintf("SELECT %s FROM %s WHERE `client_id`=? AND `email`=? LIMIT 1", frEmpRows, m.table)
	err := m.conn.QueryRowCtx(ctx, &resp, q, clientId, email)
	switch err {
	case nil:
		return &resp, nil
	case sqlx.ErrNotFound:
		return nil, ErrNotFound
	default:
		return nil, err
	}
}

func (m *customFrEmpModel) FindOneByClientIdMobile(ctx context.Context, clientId, mobile string) (*FrEmp, error) {
	clientId = strings.TrimSpace(clientId)
	mobile = strings.TrimSpace(mobile)
	if clientId == "" || mobile == "" {
		return nil, ErrNotFound
	}
	var resp FrEmp
	q := fmt.Sprintf("SELECT %s FROM %s WHERE `client_id`=? AND `mobile`=? LIMIT 1", frEmpRows, m.table)
	err := m.conn.QueryRowCtx(ctx, &resp, q, clientId, mobile)
	switch err {
	case nil:
		return &resp, nil
	case sqlx.ErrNotFound:
		return nil, ErrNotFound
	default:
		return nil, err
	}
}

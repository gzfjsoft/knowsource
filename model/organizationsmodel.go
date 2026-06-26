package model

import (
	"context"
	"fmt"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ OrganizationsModel = (*customOrganizationsModel)(nil)

type (
	// OrganizationsModel is an interface to be customized, add more methods here,
	// and implement the added methods in customOrganizationsModel.
	OrganizationsModel interface {
		organizationsModel
		withSession(session sqlx.Session) OrganizationsModel
		FindByUserId(ctx context.Context, userId uint64) ([]*Organizations, error)
		FindByOrgName(ctx context.Context, org_name string) ([]*Organizations, error)
		FindByUserIdOrgId(ctx context.Context, userId uint64, orgId uint64) (*Organizations, error)
		FindByOrgId(ctx context.Context, orgId uint64) (*Organizations, error)
		FindAllPublic(ctx context.Context) (*[]Organizations, error)
		FindAll(ctx context.Context) (*[]Organizations, error)
		FindAllPublicEx(ctx context.Context, userId uint64) (*[]OrganizationsEx, error)
		FindAllEx(ctx context.Context) (*[]OrganizationsEx, error)
		Count(ctx context.Context, condition string) (uint64, error)
		FindDefaultOne(ctx context.Context) (*Organizations, error)
	}

	customOrganizationsModel struct {
		*defaultOrganizationsModel
	}
)

type OrganizationsEx struct {
	Organizations
	Username string `db:"username"`
}

// NewOrganizationsModel returns a model for the database table.
func NewOrganizationsModel(conn sqlx.SqlConn) OrganizationsModel {
	return &customOrganizationsModel{
		defaultOrganizationsModel: newOrganizationsModel(conn),
	}
}

func (m *customOrganizationsModel) withSession(session sqlx.Session) OrganizationsModel {
	return NewOrganizationsModel(sqlx.NewSqlConnFromSession(session))
}

func (m *customOrganizationsModel) FindAll(ctx context.Context) (*[]Organizations, error) {
	var resp []Organizations
	query := fmt.Sprintf("select %s from %s", organizationsRows, m.table)
	err := m.conn.QueryRowsCtx(ctx, &resp, query)
	switch err {
	case nil:
		return &resp, nil
	case sqlx.ErrNotFound:
		return nil, ErrNotFound
	default:
		return nil, err
	}
}

func (m *customOrganizationsModel) FindAllPublic(ctx context.Context) (*[]Organizations, error) {
	var resp []Organizations
	query := fmt.Sprintf("select %s from %s  where is_private=0", organizationsRows, m.table)
	err := m.conn.QueryRowsCtx(ctx, &resp, query)
	switch err {
	case nil:
		return &resp, nil
	case sqlx.ErrNotFound:
		return nil, ErrNotFound
	default:
		return nil, err
	}
}

func (m *customOrganizationsModel) FindAllPublicEx(ctx context.Context, userId uint64) (*[]OrganizationsEx, error) {
	var resp []OrganizationsEx
	query := fmt.Sprintf(" select a.*,users.username from %s a left outer join users on  users.user_id=a.created_by  where  is_private=0 or a.created_by=?", m.table)
	err := m.conn.QueryRowsCtx(ctx, &resp, query, userId)
	switch err {
	case nil:
		return &resp, nil
	case sqlx.ErrNotFound:
		return nil, ErrNotFound
	default:
		return nil, err
	}
}

func (m *customOrganizationsModel) FindAllEx(ctx context.Context) (*[]OrganizationsEx, error) {
	var resp []OrganizationsEx
	query := fmt.Sprintf(" select a.*,users.username from %s a left outer join users on  users.user_id=a.created_by", m.table)
	err := m.conn.QueryRowsCtx(ctx, &resp, query)
	switch err {
	case nil:
		return &resp, nil
	case sqlx.ErrNotFound:
		return nil, ErrNotFound
	default:
		return nil, err
	}
}

func (m *customOrganizationsModel) FindByOrgId(ctx context.Context, orgId uint64) (*Organizations, error) {
	query := fmt.Sprintf("select %s from %s where org_id = ?", organizationsRows, m.table)
	var resp *Organizations

	err := m.conn.QueryRowCtx(ctx, &resp, query, orgId)
	switch err {
	case nil:
		return resp, nil
	case sqlx.ErrNotFound:
		return nil, ErrNotFound
	default:
		return nil, err
	}
}

func (m *customOrganizationsModel) FindByUserIdOrgId(ctx context.Context, userId uint64, orgId uint64) (*Organizations, error) {
	query := fmt.Sprintf("select %s from %s where created_by = ? and org_id = ?", organizationsRows, m.table)
	var resp Organizations

	err := m.conn.QueryRowCtx(ctx, &resp, query, userId, orgId)
	switch err {
	case nil:
		return &resp, nil
	case sqlx.ErrNotFound:
		return nil, ErrNotFound
	default:
		return nil, err
	}
}
func (m *customOrganizationsModel) FindByUserId(ctx context.Context, userId uint64) ([]*Organizations, error) {
	query := fmt.Sprintf("select %s from %s where created_by = ?", organizationsRows, m.table)
	var resp []*Organizations
	err := m.conn.QueryRowsCtx(ctx, &resp, query, userId)
	switch err {
	case nil:
		return resp, nil
	case sqlx.ErrNotFound:
		return nil, ErrNotFound
	default:
		return nil, err
	}
}

func (m *customOrganizationsModel) FindByOrgName(ctx context.Context, org_name string) ([]*Organizations, error) {
	query := fmt.Sprintf("select %s from %s where org_name = ?", organizationsRows, m.table)
	var resp []*Organizations
	err := m.conn.QueryRowsCtx(ctx, &resp, query, org_name)
	switch err {
	case nil:
		return resp, nil
	case sqlx.ErrNotFound:
		return nil, ErrNotFound
	default:
		return nil, err
	}
}

func (m *customOrganizationsModel) Count(ctx context.Context, condition string) (uint64, error) {
	var count uint64
	query := fmt.Sprintf("select count(*) from %s where %s", m.table, condition)
	err := m.conn.QueryRowCtx(ctx, &count, query)
	return count, err
}

func (m *customOrganizationsModel) FindDefaultOne(ctx context.Context) (*Organizations, error) {
	query := fmt.Sprintf("select %s from %s where `is_default` = 1 limit 1", organizationsRows, m.table)
	var resp Organizations
	err := m.conn.QueryRowCtx(ctx, &resp, query)
	switch err {
	case nil:
		return &resp, nil
	case sqlx.ErrNotFound:
		return nil, ErrNotFound
	default:
		return nil, err
	}
}

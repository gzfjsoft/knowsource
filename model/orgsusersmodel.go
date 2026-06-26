package model

import (
	"context"
	"fmt"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ OrgsUsersModel = (*customOrgsUsersModel)(nil)

type (
	// OrgsUsersModel is an interface to be customized, add more methods here,
	// and implement the added methods in customOrgsUsersModel.
	OrgsUsersModel interface {
		orgsUsersModel
		WithSession(session sqlx.Session) OrgsUsersModel
		FindAllByUserId(ctx context.Context, userId uint64) (*[]Org, error)
		FindAllByOrgId(ctx context.Context, OrgId uint64) (*[]Org, error)
		// FindAllPublic(ctx context.Context) (*[]Org, error)

		FindOneByOrgIdUserId(ctx context.Context, orgId uint64, userId uint64) (*OrgsUsers, error)
		DeleteExOwner(ctx context.Context, uid uint64, oid uint64) error
		DeleteExNotOwner(ctx context.Context, uid uint64, oid uint64) error
		// GetUserOrgs(ctx context.Context, userId int64) (*[]Org, error)
		DeleteByUserId(ctx context.Context, userId uint64) error
	}

	customOrgsUsersModel struct {
		*defaultOrgsUsersModel
	}
)

type Org struct {
	UserId  uint64 `json:"userId"`
	Role    string `json:"role"`
	OrgId   uint64 `json:"orgId"`
	OrgName string `json:"orgName"`
}

// NewOrgsUsersModel returns a model for the database table.
func NewOrgsUsersModel(conn sqlx.SqlConn) OrgsUsersModel {
	return &customOrgsUsersModel{
		defaultOrgsUsersModel: newOrgsUsersModel(conn),
	}
}

func (m *customOrgsUsersModel) DeleteByUserId(ctx context.Context, userId uint64) error {
	query := fmt.Sprintf("delete from %s where `user_id` = ?", m.table)
	_, err := m.conn.ExecCtx(ctx, query, userId)
	return err
}

func (m *customOrgsUsersModel) WithSession(session sqlx.Session) OrgsUsersModel {
	return NewOrgsUsersModel(sqlx.NewSqlConnFromSession(session))
}

func (m *customOrgsUsersModel) DeleteExNotOwner(ctx context.Context, uid uint64, oid uint64) error {
	query := fmt.Sprintf("delete from %s where `user_id` = ? and `org_id` = ? and role  != 'owner'", m.table)
	_, err := m.conn.ExecCtx(ctx, query, uid, oid)
	return err
}

func (m *customOrgsUsersModel) DeleteExOwner(ctx context.Context, uid uint64, oid uint64) error {
	query := fmt.Sprintf("delete from %s where `user_id` = ? and `org_id` = ? and role  = 'owner'", m.table)
	_, err := m.conn.ExecCtx(ctx, query, uid, oid)
	return err
}

func (m *customOrgsUsersModel) FindAllByUserId(ctx context.Context, userId uint64) (*[]Org, error) {
	query := fmt.Sprintf("select orgs_users.user_id, orgs_users.role, orgs_users.org_id, organizations.org_name from `orgs_users`,organizations where orgs_users.org_id=organizations.org_id and  orgs_users.user_id=%d", userId)
	var resp []Org
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

func (m *customOrgsUsersModel) FindAllByOrgId(ctx context.Context, OrgId uint64) (*[]Org, error) {
	var resp []Org
	query := "select orgs_users.user_id, orgs_users.role, orgs_users.org_id, organizations.org_name from `orgs_users`,organizations where orgs_users.org_id=organizations.org_id and  orgs_users.org_id=?"

	//	query := fmt.Sprintf("select %s from %s where `org_id` =?", orgsUsersRows, m.table)
	err := m.conn.QueryRowsCtx(ctx, &resp, query, OrgId)
	switch err {
	case nil:
		return &resp, nil
	case sqlx.ErrNotFound:
		return nil, ErrNotFound
	default:
		return nil, err
	}
}

// func (m *customOrgsUsersModel) FindAllByUserId(ctx context.Context, userId uint64) (*[]Org, error) {
// 	var resp []Org
// 	query := fmt.Sprintf("select %s from %s where `user_id` =?", orgsUsersRows, m.table)
// 	err := m.conn.QueryRowsCtx(ctx, &resp, query, userId)
// 	switch err {
// 	case nil:
// 		return &resp, nil
// 	case sqlx.ErrNotFound:
// 		return nil, ErrNotFound
// 	default:
// 		return nil, err
// 	}
// }

func (m *customOrgsUsersModel) FindOneByOrgIdUserId(ctx context.Context, orgId uint64, userId uint64) (*OrgsUsers, error) {
	var resp OrgsUsers
	query := fmt.Sprintf("select %s from %s where `org_id` = ? and `user_id` = ? limit 1", orgsUsersRows, m.table)
	err := m.conn.QueryRowCtx(ctx, &resp, query, orgId, userId)
	switch err {
	case nil:
		return &resp, nil
	case sqlx.ErrNotFound:
		return nil, ErrNotFound
	default:
		return nil, err
	}
}

// func (m *customOrgsUsersModel) GetUserOrgs(ctx context.Context, userId int64) (*[]Org, error) {
// 	var resp []Org
// 	query := "select orgs_users.user_id, orgs_users.role, orgs_users.org_id, organizations.org_name from `orgs_users`,organizations where orgs_users.org_id=organizations.org_id and  orgs_users.user_id=?"

// 	//	query := fmt.Sprintf("select %s from %s where `org_id` =?", orgsUsersRows, m.table)
// 	err := m.conn.QueryRowsCtx(ctx, &resp, query, userId)
// 	switch err {
// 	case nil:
// 		return &resp, nil
// 	case sqlx.ErrNotFound:
// 		return nil, ErrNotFound
// 	default:
// 		return nil, err
// 	}
// }

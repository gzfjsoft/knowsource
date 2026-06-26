package model

import (
	"context"
	"fmt"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ UsersUpdateModel = (*customUsersUpdateModel)(nil)

type (
	// UsersUpdateModel is an interface to be customized, add more methods here,
	// and implement the added methods in customUsersUpdateModel.
	UsersUpdateModel interface {
		usersUpdateModel
		withSession(session sqlx.Session) UsersUpdateModel
		FindAllWithPagination(ctx context.Context, page, pageSize uint64) ([]*UsersUpdate, error)
		Count(ctx context.Context) (uint64, error)
	}

	customUsersUpdateModel struct {
		*defaultUsersUpdateModel
	}
)

// NewUsersUpdateModel returns a model for the database table.
func NewUsersUpdateModel(conn sqlx.SqlConn) UsersUpdateModel {
	return &customUsersUpdateModel{
		defaultUsersUpdateModel: newUsersUpdateModel(conn),
	}
}

func (m *customUsersUpdateModel) withSession(session sqlx.Session) UsersUpdateModel {
	return NewUsersUpdateModel(sqlx.NewSqlConnFromSession(session))
}

// FindAllWithPagination returns all users with pagination
func (m *customUsersUpdateModel) FindAllWithPagination(ctx context.Context, page, pageSize uint64) ([]*UsersUpdate, error) {
	query := fmt.Sprintf("SELECT %s FROM %s ORDER BY updated_at DESC LIMIT ? OFFSET ?", usersUpdateRows, m.table)
	offset := (page - 1) * pageSize

	var resp []*UsersUpdate
	err := m.conn.QueryRowsCtx(ctx, &resp, query, pageSize, offset)
	return resp, err
}

// Count returns the total number of users
func (m *customUsersUpdateModel) Count(ctx context.Context) (uint64, error) {
	query := fmt.Sprintf("SELECT COUNT(*) FROM %s", m.table)
	var count uint64
	err := m.conn.QueryRowCtx(ctx, &count, query)
	return count, err
}

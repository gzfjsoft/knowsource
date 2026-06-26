package model

import (
	"context"
	"fmt"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ UserAcquireUserModel = (*customUserAcquireUserModel)(nil)

type (
	// UserAcquireUserModel is an interface to be customized, add more methods here,
	// and implement the added methods in customUserAcquireUserModel.
	UserAcquireUserModel interface {
		userAcquireUserModel
		withSession(session sqlx.Session) UserAcquireUserModel
		FindAllByUserId(ctx context.Context, userId uint64) ([]*UserAcquireUser, error)
	}

	customUserAcquireUserModel struct {
		*defaultUserAcquireUserModel
	}
)

// NewUserAcquireUserModel returns a model for the database table.
func NewUserAcquireUserModel(conn sqlx.SqlConn) UserAcquireUserModel {
	return &customUserAcquireUserModel{
		defaultUserAcquireUserModel: newUserAcquireUserModel(conn),
	}
}

func (m *customUserAcquireUserModel) withSession(session sqlx.Session) UserAcquireUserModel {
	return NewUserAcquireUserModel(sqlx.NewSqlConnFromSession(session))
}

func (m *customUserAcquireUserModel) FindAllByUserId(ctx context.Context, userId uint64) ([]*UserAcquireUser, error) {
	query := fmt.Sprintf("SELECT * FROM %s WHERE user_id = ?", m.table)
	var resp []*UserAcquireUser
	err := m.conn.QueryRowsCtx(ctx, &resp, query, userId)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

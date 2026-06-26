package model

import (
	"context"
	"fmt"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ UserAuthIdcardModel = (*customUserAuthIdcardModel)(nil)

type (
	// UserAuthIdcardModel is an interface to be customized, add more methods here,
	// and implement the added methods in customUserAuthIdcardModel.
	UserAuthIdcardModel interface {
		userAuthIdcardModel
		withSession(session sqlx.Session) UserAuthIdcardModel
		GetByUserId(ctx context.Context, userId int64) (*UserAuthIdcard, error)
		GetVerifyIdcardList(ctx context.Context, page int64, pageSize int64) ([]*UserAuthIdcard, error)
	}

	customUserAuthIdcardModel struct {
		*defaultUserAuthIdcardModel
	}
)

// NewUserAuthIdcardModel returns a model for the database table.
func NewUserAuthIdcardModel(conn sqlx.SqlConn) UserAuthIdcardModel {
	return &customUserAuthIdcardModel{
		defaultUserAuthIdcardModel: newUserAuthIdcardModel(conn),
	}
}

func (m *customUserAuthIdcardModel) withSession(session sqlx.Session) UserAuthIdcardModel {
	return NewUserAuthIdcardModel(sqlx.NewSqlConnFromSession(session))
}

func (m *customUserAuthIdcardModel) GetByUserId(ctx context.Context, userId int64) (*UserAuthIdcard, error) {
	query := fmt.Sprintf("select %s from %s where `user_id` = ? and audit_status=1 order by id desc limit 1", userAuthIdcardRows, m.table)
	var resp UserAuthIdcard
	err := m.conn.QueryRowCtx(ctx, &resp, query, userId)
	switch err {
	case nil:
		return &resp, nil
	case sqlx.ErrNotFound:
		return nil, ErrNotFound
	default:
		return nil, err
	}
}

func (m *customUserAuthIdcardModel) GetVerifyIdcardList(ctx context.Context, page int64, pageSize int64) ([]*UserAuthIdcard, error) {
	query := fmt.Sprintf("select %s from %s where audit_status=1 order by id desc limit %d, %d", userAuthIdcardRows, m.table, (page-1)*pageSize, pageSize)
	var resp []*UserAuthIdcard
	err := m.conn.QueryRowsCtx(ctx, &resp, query)
	return resp, err
}

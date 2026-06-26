package model

import (
	"context"
	"fmt"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ ApplyJoinModel = (*customApplyJoinModel)(nil)

type (
	// ApplyJoinModel is an interface to be customized, add more methods here,
	// and implement the added methods in customApplyJoinModel.
	ApplyJoinModel interface {
		applyJoinModel
		withSession(session sqlx.Session) ApplyJoinModel
		FindAll(ctx context.Context, uid int64, page, pageSize int64) (*[]ApplyJoin, error)
		Count(ctx context.Context, uid int64) (int64, error)
	}

	customApplyJoinModel struct {
		*defaultApplyJoinModel
	}
)

// NewApplyJoinModel returns a model for the database table.
func NewApplyJoinModel(conn sqlx.SqlConn) ApplyJoinModel {
	return &customApplyJoinModel{
		defaultApplyJoinModel: newApplyJoinModel(conn),
	}
}

func (m *customApplyJoinModel) withSession(session sqlx.Session) ApplyJoinModel {
	return NewApplyJoinModel(sqlx.NewSqlConnFromSession(session))
}

func (m *defaultApplyJoinModel) FindAll(ctx context.Context, uid int64, page, pageSize int64) (*[]ApplyJoin, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}
	offset := (page - 1) * pageSize

	query := fmt.Sprintf("select %s from %s where user_id = ? limit ?,?", applyJoinRows, m.table)
	var resp []ApplyJoin
	err := m.conn.QueryRowsCtx(ctx, &resp, query, uid, offset, pageSize)
	switch err {
	case nil:
		return &resp, nil
	case sqlx.ErrNotFound:
		return nil, ErrNotFound
	default:
		return nil, err
	}
}

func (m *defaultApplyJoinModel) Count(ctx context.Context, uid int64) (int64, error) {
	query := fmt.Sprintf("select count(*) from %s where user_id = ?", m.table)
	var count int64
	err := m.conn.QueryRowCtx(ctx, &count, query, uid)
	switch err {
	case nil:
		return count, nil
	default:
		return 0, err
	}
}

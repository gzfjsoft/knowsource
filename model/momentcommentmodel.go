package model

import (
	"context"
	"fmt"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ MomentCommentModel = (*customMomentCommentModel)(nil)

type (
	// MomentCommentModel is an interface to be customized, add more methods here,
	// and implement the added methods in customMomentCommentModel.
	MomentCommentModel interface {
		momentCommentModel
		withSession(session sqlx.Session) MomentCommentModel
		FindByConditions(ctx context.Context, condition string, values ...interface{}) (*[]MomentComment, error)
		FindMomentAll(ctx context.Context, momentId uint64) (*[]MomentComment, error)
		CountByMomentId(ctx context.Context, momentId uint64) (uint64, error)
	}

	customMomentCommentModel struct {
		*defaultMomentCommentModel
	}
)

// NewMomentCommentModel returns a model for the database table.
func NewMomentCommentModel(conn sqlx.SqlConn) MomentCommentModel {
	return &customMomentCommentModel{
		defaultMomentCommentModel: newMomentCommentModel(conn),
	}
}

func (m *customMomentCommentModel) withSession(session sqlx.Session) MomentCommentModel {
	return NewMomentCommentModel(sqlx.NewSqlConnFromSession(session))
}
func (m *customMomentCommentModel) FindMomentAll(ctx context.Context, momentId uint64) (*[]MomentComment, error) {
	query := fmt.Sprintf("select %s from %s where moment_id = ?", momentCommentRows, m.table)
	var resp []MomentComment
	err := m.conn.QueryRowsCtx(ctx, &resp, query, momentId)
	switch err {
	case nil:
		return &resp, nil
	default:
		return nil, err
	}
}
func (m *customMomentCommentModel) FindByConditions(ctx context.Context, condition string, values ...interface{}) (*[]MomentComment, error) {
	query := fmt.Sprintf("select %s from %s where %s", momentCommentRows, m.table, condition)
	var resp []MomentComment
	err := m.conn.QueryRowsCtx(ctx, &resp, query, values...)
	switch err {
	case nil:
		return &resp, nil
	default:
		return nil, err
	}
}

func (m *customMomentCommentModel) CountByMomentId(ctx context.Context, momentId uint64) (uint64, error) {
	query := fmt.Sprintf("select count(*) from %s where moment_id = ?", m.table)
	var count uint64
	err := m.conn.QueryRowCtx(ctx, &count, query, momentId)
	return count, err
}

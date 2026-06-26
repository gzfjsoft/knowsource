package model

import (
	"context"
	"fmt"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ MomentLikeModel = (*customMomentLikeModel)(nil)

type (
	// MomentLikeModel is an interface to be customized, add more methods here,
	// and implement the added methods in customMomentLikeModel.
	MomentLikeModel interface {
		momentLikeModel
		withSession(session sqlx.Session) MomentLikeModel
		FindAllByMomentId(ctx context.Context, momentId int64) ([]*MomentLike, error)
	}

	customMomentLikeModel struct {
		*defaultMomentLikeModel
	}
)

// NewMomentLikeModel returns a model for the database table.
func NewMomentLikeModel(conn sqlx.SqlConn) MomentLikeModel {
	return &customMomentLikeModel{
		defaultMomentLikeModel: newMomentLikeModel(conn),
	}
}

func (m *customMomentLikeModel) withSession(session sqlx.Session) MomentLikeModel {
	return NewMomentLikeModel(sqlx.NewSqlConnFromSession(session))
}

func (m *customMomentLikeModel) FindAllByMomentId(ctx context.Context, momentId int64) ([]*MomentLike, error) {
	query := fmt.Sprintf("select %s from %s where `moment_id` = ?", momentLikeRows, m.table)
	var resp []*MomentLike
	err := m.conn.QueryRowsCtx(ctx, &resp, query, momentId)
	return resp, err
}

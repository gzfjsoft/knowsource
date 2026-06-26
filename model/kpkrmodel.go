package model

import (
	"context"
	"fmt"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ KpKrModel = (*customKpKrModel)(nil)

type (
	// KpKrModel is an interface to be customized, add more methods here,
	// and implement the added methods in customKpKrModel.
	KpKrModel interface {
		kpKrModel
		withSession(session sqlx.Session) KpKrModel
		Count(ctx context.Context, kpId, krId uint64) (int64, error)
		FindAll(ctx context.Context, page, pageSize int, kpId, krId uint64) ([]*KpKr, error)
	}

	customKpKrModel struct {
		*defaultKpKrModel
	}
)

// NewKpKrModel returns a model for the database table.
func NewKpKrModel(conn sqlx.SqlConn) KpKrModel {
	return &customKpKrModel{
		defaultKpKrModel: newKpKrModel(conn),
	}
}

func (m *customKpKrModel) withSession(session sqlx.Session) KpKrModel {
	return NewKpKrModel(sqlx.NewSqlConnFromSession(session))
}

func (m *customKpKrModel) Count(ctx context.Context, kpId, krId uint64) (int64, error) {
	var count int64
	query := fmt.Sprintf("select count(*) from %s", m.table)
	args := []interface{}{}

	// Build WHERE clause based on provided filters
	if kpId > 0 || krId > 0 {
		query += " where"
		if kpId > 0 {
			query += " kp_id = ?"
			args = append(args, kpId)
		}
		if krId > 0 {
			if kpId > 0 {
				query += " and"
			}
			query += " kr_id = ?"
			args = append(args, krId)
		}
	}

	err := m.conn.QueryRowCtx(ctx, &count, query, args...)
	return count, err
}

func (m *customKpKrModel) FindAll(ctx context.Context, page, pageSize int, kpId, krId uint64) ([]*KpKr, error) {
	query := fmt.Sprintf("select %s from %s", kpKrRows, m.table)
	args := []interface{}{}

	// Build WHERE clause based on provided filters
	if kpId > 0 || krId > 0 {
		query += " where"
		if kpId > 0 {
			query += " kp_id = ?"
			args = append(args, kpId)
		}
		if krId > 0 {
			if kpId > 0 {
				query += " and"
			}
			query += " kr_id = ?"
			args = append(args, krId)
		}
	}

	query += " limit ? offset ?"
	args = append(args, pageSize, (page-1)*pageSize)

	var resp []*KpKr
	err := m.conn.QueryRowsCtx(ctx, &resp, query, args...)
	return resp, err
}

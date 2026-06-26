package model

import (
	"context"
	"fmt"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ KpPrerequisiteModel = (*customKpPrerequisiteModel)(nil)

type (
	// KpPrerequisiteModel is an interface to be customized, add more methods here,
	// and implement the added methods in customKpPrerequisiteModel.
	KpPrerequisiteModel interface {
		kpPrerequisiteModel
		withSession(session sqlx.Session) KpPrerequisiteModel
		Count(ctx context.Context, kpId uint64) (int64, error)
		FindAll(ctx context.Context, page, pageSize int, kpId uint64) ([]*KpPrerequisite, error)
	}

	customKpPrerequisiteModel struct {
		*defaultKpPrerequisiteModel
	}
)

// NewKpPrerequisiteModel returns a model for the database table.
func NewKpPrerequisiteModel(conn sqlx.SqlConn) KpPrerequisiteModel {
	return &customKpPrerequisiteModel{
		defaultKpPrerequisiteModel: newKpPrerequisiteModel(conn),
	}
}

func (m *customKpPrerequisiteModel) withSession(session sqlx.Session) KpPrerequisiteModel {
	return NewKpPrerequisiteModel(sqlx.NewSqlConnFromSession(session))
}

func (m *customKpPrerequisiteModel) Count(ctx context.Context, kpId uint64) (int64, error) {
	var count int64
	query := fmt.Sprintf("select count(*) from %s", m.table)
	args := []interface{}{}

	if kpId != 0 {
		query += " where parent_kp_id = ? or child_kp_id = ?"
		args = append(args, kpId, kpId)
	}

	err := m.conn.QueryRowCtx(ctx, &count, query, args...)
	return count, err
}

func (m *customKpPrerequisiteModel) FindAll(ctx context.Context, page, pageSize int, kpId uint64) ([]*KpPrerequisite, error) {
	query := fmt.Sprintf("select %s from %s", kpPrerequisiteRows, m.table)
	args := []interface{}{}

	if kpId != 0 {
		query += " where parent_kp_id = ? or child_kp_id = ?"
		args = append(args, kpId, kpId)
	}

	query += " limit ? offset ?"
	args = append(args, pageSize, (page-1)*pageSize)

	var resp []*KpPrerequisite
	err := m.conn.QueryRowsCtx(ctx, &resp, query, args...)
	return resp, err
}

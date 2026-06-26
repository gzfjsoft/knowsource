package model

import (
	"context"
	"fmt"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ KnowledgePointModel = (*customKnowledgePointModel)(nil)

type (
	// KnowledgePointModel is an interface to be customized, add more methods here,
	// and implement the added methods in customKnowledgePointModel.
	KnowledgePointModel interface {
		knowledgePointModel
		withSession(session sqlx.Session) KnowledgePointModel
		Count(ctx context.Context, keyword string) (int64, error)
		FindAll(ctx context.Context, page, pageSize int, keyword string) ([]*KnowledgePoint, error)
	}

	customKnowledgePointModel struct {
		*defaultKnowledgePointModel
	}
)

// NewKnowledgePointModel returns a model for the database table.
func NewKnowledgePointModel(conn sqlx.SqlConn) KnowledgePointModel {
	return &customKnowledgePointModel{
		defaultKnowledgePointModel: newKnowledgePointModel(conn),
	}
}

func (m *customKnowledgePointModel) withSession(session sqlx.Session) KnowledgePointModel {
	return NewKnowledgePointModel(sqlx.NewSqlConnFromSession(session))
}

func (m *customKnowledgePointModel) Count(ctx context.Context, keyword string) (int64, error) {
	var count int64
	query := fmt.Sprintf("select count(*) from %s", m.table)
	args := []interface{}{}

	if keyword != "" {
		query += " where name like ? or description like ?"
		args = append(args, "%"+keyword+"%", "%"+keyword+"%")
	}

	err := m.conn.QueryRowCtx(ctx, &count, query, args...)
	return count, err
}

func (m *customKnowledgePointModel) FindAll(ctx context.Context, page, pageSize int, keyword string) ([]*KnowledgePoint, error) {
	query := fmt.Sprintf("select %s from %s", knowledgePointRows, m.table)
	args := []interface{}{}

	if keyword != "" {
		query += " where name like ? or description like ?"
		args = append(args, "%"+keyword+"%", "%"+keyword+"%")
	}

	query += " limit ? offset ?"
	args = append(args, pageSize, (page-1)*pageSize)

	var resp []*KnowledgePoint
	err := m.conn.QueryRowsCtx(ctx, &resp, query, args...)
	return resp, err
}

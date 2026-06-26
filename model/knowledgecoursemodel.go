package model

import (
	"context"
	"fmt"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ KnowledgeCourseModel = (*customKnowledgeCourseModel)(nil)

type (
	// KnowledgeCourseModel is an interface to be customized, add more methods here,
	// and implement the added methods in customKnowledgeCourseModel.
	KnowledgeCourseModel interface {
		knowledgeCourseModel
		withSession(session sqlx.Session) KnowledgeCourseModel
		FindAll(ctx context.Context, page, pageSize int, keyword string) ([]*KnowledgeCourse, int64, error)
	}

	customKnowledgeCourseModel struct {
		*defaultKnowledgeCourseModel
	}
)

// NewKnowledgeCourseModel returns a model for the database table.
func NewKnowledgeCourseModel(conn sqlx.SqlConn) KnowledgeCourseModel {
	return &customKnowledgeCourseModel{
		defaultKnowledgeCourseModel: newKnowledgeCourseModel(conn),
	}
}

func (m *customKnowledgeCourseModel) withSession(session sqlx.Session) KnowledgeCourseModel {
	return NewKnowledgeCourseModel(sqlx.NewSqlConnFromSession(session))
}

func (m *customKnowledgeCourseModel) FindAll(ctx context.Context, page, pageSize int, keyword string) ([]*KnowledgeCourse, int64, error) {
	var courses []*KnowledgeCourse
	var count int64

	whereClause := ""
	args := []interface{}{}

	if keyword != "" {
		whereClause = "WHERE name LIKE ? OR category LIKE ? OR description LIKE ?"
		keyword = "%" + keyword + "%"
		args = append(args, keyword, keyword, keyword)
	}

	// Get total count
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM %s %s", m.table, whereClause)
	err := m.conn.QueryRowCtx(ctx, &count, countQuery, args...)
	if err != nil {
		return nil, 0, err
	}

	// Get paginated data
	query := fmt.Sprintf("SELECT %s FROM %s %s ORDER BY sort_order LIMIT ?, ?", knowledgeCourseRows, m.table, whereClause)
	args = append(args, (page-1)*pageSize, pageSize)
	err = m.conn.QueryRowsCtx(ctx, &courses, query, args...)
	if err != nil {
		return nil, 0, err
	}

	return courses, count, nil
}

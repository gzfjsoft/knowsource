package model

import (
	"context"
	"fmt"
	"strings"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"github.com/zeromicro/go-zero/core/stringx"
)

var _ ExamsModel = (*customExamsModel)(nil)

type (
	// ExamsModel is an interface to be customized, add more methods here,
	// and implement the added methods in customExamsModel.
	ExamsModel interface {
		examsModel
		withSession(session sqlx.Session) ExamsModel
		FindAll(ctx context.Context, conditions []string, args []interface{}, limit string) ([]*Exams, error)
		FindAllLite(ctx context.Context, conditions []string, args []interface{}, limit string) ([]*Exams, error)
		Count(ctx context.Context, conditions []string, args []interface{}) (uint64, error)
		FindOneBySname(ctx context.Context, sname string) (*Exams, error)
	}

	customExamsModel struct {
		*defaultExamsModel
	}
)

// NewExamsModel returns a model for the database table.
func NewExamsModel(conn sqlx.SqlConn) ExamsModel {
	return &customExamsModel{
		defaultExamsModel: newExamsModel(conn),
	}
}

func (m *customExamsModel) withSession(session sqlx.Session) ExamsModel {
	return NewExamsModel(sqlx.NewSqlConnFromSession(session))
}

func (m *customExamsModel) FindOneBySname(ctx context.Context, sname string) (*Exams, error) {
	query := fmt.Sprintf("select %s from %s where exam_sname = ? limit 1", examsRows, m.table)
	var resp Exams
	err := m.conn.QueryRowCtx(ctx, &resp, query, sname)
	switch err {
	case nil:
		return &resp, nil
	case sqlx.ErrNotFound:
		return nil, ErrNotFound
	default:
		return nil, err
	}
}

func (m *customExamsModel) FindAll(ctx context.Context, conditions []string, args []interface{}, limit string) ([]*Exams, error) {

	where_str := ""
	if len(conditions) > 0 {
		where_str += " where 1=1 and " + strings.Join(conditions, " and ")
	}

	query := fmt.Sprintf("select %s from %s %s order by sort_order %s", examsRows, m.table, where_str, limit)
	logx.Infof("query: %s", query)

	var resp []*Exams
	err := m.conn.QueryRowsCtx(ctx, &resp, query, args...)
	switch err {
	case nil:
		return resp, nil
	case sqlx.ErrNotFound:
		return nil, ErrNotFound
	default:
		return nil, err
	}
}

func (m *customExamsModel) FindAllLite(ctx context.Context, conditions []string, args []interface{}, limit string) ([]*Exams, error) {

	where_str := ""
	if len(conditions) > 0 {
		where_str += " where 1=1 and " + strings.Join(conditions, " and ")
	}

	examsRowsExpectAutoSet = strings.Join(stringx.Remove(examsFieldNames, "`questions`"), ",")

	query := fmt.Sprintf("select %s from %s %s order by sort_order %s", examsRowsExpectAutoSet, m.table, where_str, limit)
	logx.Infof("query: %s", query)

	var resp []*Exams
	err := m.conn.QueryRowsCtx(ctx, &resp, query, args...)
	switch err {
	case nil:
		return resp, nil
	case sqlx.ErrNotFound:
		return nil, ErrNotFound
	default:
		return nil, err
	}
}

func (m *customExamsModel) Count(ctx context.Context, conditions []string, args []interface{}) (uint64, error) {
	query := fmt.Sprintf("select count(*) from %s", m.table)
	if len(conditions) > 0 {
		query += " where " + strings.Join(conditions, " and ")
	}

	var count uint64
	err := m.conn.QueryRowCtx(ctx, &count, query, args...)
	switch err {
	case nil:
		return count, nil
	case sqlx.ErrNotFound:
		return 0, ErrNotFound
	default:
		return 0, err
	}
}

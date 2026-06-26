package model

import (
	"context"
	"fmt"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ ExamsResultModel = (*customExamsResultModel)(nil)

type (
	// ExamsResultModel is an interface to be customized, add more methods here,
	// and implement the added methods in customExamsResultModel.
	ExamsResultModel interface {
		examsResultModel
		withSession(session sqlx.Session) ExamsResultModel
		FindAllByUserIdAndExamId(ctx context.Context, userId, examId uint64) ([]*ExamsResult, error)
		FindAllByExamId(ctx context.Context, examId uint64) ([]*ExamsResult, error)
		FindTempOneByUserIdAndExamId(ctx context.Context, userId, examId uint64) (*ExamsResult, error)
	}

	customExamsResultModel struct {
		*defaultExamsResultModel
	}
)

// NewExamsResultModel returns a model for the database table.
func NewExamsResultModel(conn sqlx.SqlConn) ExamsResultModel {
	return &customExamsResultModel{
		defaultExamsResultModel: newExamsResultModel(conn),
	}
}

func (m *customExamsResultModel) withSession(session sqlx.Session) ExamsResultModel {
	return NewExamsResultModel(sqlx.NewSqlConnFromSession(session))
}

func (m *customExamsResultModel) FindTempOneByUserIdAndExamId(ctx context.Context, userId, examId uint64) (*ExamsResult, error) {
	query := fmt.Sprintf("select %s from %s where `user_id` = ? and `exam_id` = ? and `is_temp` = 1 order by created_at desc", examsResultRows, m.table)
	var resp ExamsResult
	err := m.conn.QueryRowCtx(ctx, &resp, query, userId, examId)
	switch err {
	case nil:
		return &resp, nil
	case sqlx.ErrNotFound:
		return nil, ErrNotFound
	default:
		return nil, err
	}
}

func (m *customExamsResultModel) FindAllByUserIdAndExamId(ctx context.Context, userId, examId uint64) ([]*ExamsResult, error) {
	query := fmt.Sprintf("select %s from %s where `user_id` = ? and `exam_id` = ?   order by created_at desc", examsResultRows, m.table)
	var resp []*ExamsResult
	err := m.conn.QueryRowsCtx(ctx, &resp, query, userId, examId)
	switch err {
	case nil:
		return resp, nil
	case sqlx.ErrNotFound:
		return nil, ErrNotFound
	default:
		return nil, err
	}
}

func (m *customExamsResultModel) FindAllByExamId(ctx context.Context, examId uint64) ([]*ExamsResult, error) {
	query := fmt.Sprintf("select %s from %s where `exam_id` = ? order by created_at desc", examsResultRows, m.table)
	var resp []*ExamsResult
	err := m.conn.QueryRowsCtx(ctx, &resp, query, examId)
	switch err {
	case nil:
		return resp, nil
	case sqlx.ErrNotFound:
		return nil, ErrNotFound
	default:
		return nil, err
	}
}

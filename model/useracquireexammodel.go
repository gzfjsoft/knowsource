package model

import (
	"context"
	"fmt"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ UserAcquireExamModel = (*customUserAcquireExamModel)(nil)

type (
	// UserAcquireExamModel is an interface to be customized, add more methods here,
	// and implement the added methods in customUserAcquireExamModel.
	UserAcquireExamModel interface {
		userAcquireExamModel
		withSession(session sqlx.Session) UserAcquireExamModel
		FindAllByUserId(ctx context.Context, userId uint64) (*[]UserAcquireExam, error)
		FindOneByUserIdAndExamId(ctx context.Context, userId uint64, examId uint64) (*UserAcquireExam, error)
	}

	customUserAcquireExamModel struct {
		*defaultUserAcquireExamModel
	}
)

// NewUserAcquireExamModel returns a model for the database table.
func NewUserAcquireExamModel(conn sqlx.SqlConn) UserAcquireExamModel {
	return &customUserAcquireExamModel{
		defaultUserAcquireExamModel: newUserAcquireExamModel(conn),
	}
}

func (m *customUserAcquireExamModel) withSession(session sqlx.Session) UserAcquireExamModel {
	return NewUserAcquireExamModel(sqlx.NewSqlConnFromSession(session))
}

func (m *customUserAcquireExamModel) FindAllByUserId(ctx context.Context, userId uint64) (*[]UserAcquireExam, error) {
	query := fmt.Sprintf("select %s from %s where `user_id` = ?", userAcquireExamRows, m.table)
	var resp []UserAcquireExam
	err := m.conn.QueryRowsCtx(ctx, &resp, query, userId)
	switch err {
	case nil:
		return &resp, nil
	case sqlx.ErrNotFound:
		return nil, ErrNotFound
	default:
		return nil, err
	}

}

func (m *customUserAcquireExamModel) FindOneByUserIdAndExamId(ctx context.Context, userId uint64, examId uint64) (*UserAcquireExam, error) {
	query := fmt.Sprintf("select %s from %s where `user_id` = ? and `acq_exam_id` = ?", userAcquireExamRows, m.table)
	var resp UserAcquireExam
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

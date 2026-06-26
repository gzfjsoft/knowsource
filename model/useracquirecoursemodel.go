package model

import (
	"context"
	"fmt"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ UserAcquireCourseModel = (*customUserAcquireCourseModel)(nil)

type (
	// UserAcquireCourseModel is an interface to be customized, add more methods here,
	// and implement the added methods in customUserAcquireCourseModel.
	UserAcquireCourseModel interface {
		userAcquireCourseModel
		withSession(session sqlx.Session) UserAcquireCourseModel
		FindAllByUserId(ctx context.Context, userId uint64) (*[]UserAcquireCourse, error)
		FindOneByUserIdAndCourseId(ctx context.Context, userId uint64, courseId uint64) (*UserAcquireCourse, error)
	}

	customUserAcquireCourseModel struct {
		*defaultUserAcquireCourseModel
	}
)

// NewUserAcquireCourseModel returns a model for the database table.
func NewUserAcquireCourseModel(conn sqlx.SqlConn) UserAcquireCourseModel {
	return &customUserAcquireCourseModel{
		defaultUserAcquireCourseModel: newUserAcquireCourseModel(conn),
	}
}

func (m *customUserAcquireCourseModel) withSession(session sqlx.Session) UserAcquireCourseModel {
	return NewUserAcquireCourseModel(sqlx.NewSqlConnFromSession(session))
}

func (m *customUserAcquireCourseModel) FindAllByUserId(ctx context.Context, userId uint64) (*[]UserAcquireCourse, error) {
	query := fmt.Sprintf("select %s from %s where `user_id` = ?", userAcquireCourseRows, m.table)
	var resp []UserAcquireCourse
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

func (m *customUserAcquireCourseModel) FindOneByUserIdAndCourseId(ctx context.Context, userId uint64, courseId uint64) (*UserAcquireCourse, error) {
	query := fmt.Sprintf("select %s from %s where `user_id` = ? and `acq_course_id` = ?", userAcquireCourseRows, m.table)
	var resp UserAcquireCourse
	err := m.conn.QueryRowCtx(ctx, &resp, query, userId, courseId)
	switch err {
	case nil:
		return &resp, nil
	case sqlx.ErrNotFound:
		return nil, ErrNotFound
	default:
		return nil, err
	}
}

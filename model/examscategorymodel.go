package model

import "github.com/zeromicro/go-zero/core/stores/sqlx"

var _ ExamsCategoryModel = (*customExamsCategoryModel)(nil)

type (
	// ExamsCategoryModel is an interface to be customized, add more methods here,
	// and implement the added methods in customExamsCategoryModel.
	ExamsCategoryModel interface {
		examsCategoryModel
		withSession(session sqlx.Session) ExamsCategoryModel
	}

	customExamsCategoryModel struct {
		*defaultExamsCategoryModel
	}
)

// NewExamsCategoryModel returns a model for the database table.
func NewExamsCategoryModel(conn sqlx.SqlConn) ExamsCategoryModel {
	return &customExamsCategoryModel{
		defaultExamsCategoryModel: newExamsCategoryModel(conn),
	}
}

func (m *customExamsCategoryModel) withSession(session sqlx.Session) ExamsCategoryModel {
	return NewExamsCategoryModel(sqlx.NewSqlConnFromSession(session))
}

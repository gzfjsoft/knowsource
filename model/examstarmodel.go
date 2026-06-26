package model

import "github.com/zeromicro/go-zero/core/stores/sqlx"

var _ ExamStarModel = (*customExamStarModel)(nil)

type (
	// ExamStarModel is an interface to be customized, add more methods here,
	// and implement the added methods in customExamStarModel.
	ExamStarModel interface {
		examStarModel
		withSession(session sqlx.Session) ExamStarModel
	}

	customExamStarModel struct {
		*defaultExamStarModel
	}
)

// NewExamStarModel returns a model for the database table.
func NewExamStarModel(conn sqlx.SqlConn) ExamStarModel {
	return &customExamStarModel{
		defaultExamStarModel: newExamStarModel(conn),
	}
}

func (m *customExamStarModel) withSession(session sqlx.Session) ExamStarModel {
	return NewExamStarModel(sqlx.NewSqlConnFromSession(session))
}

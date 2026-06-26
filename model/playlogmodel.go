package model

import "github.com/zeromicro/go-zero/core/stores/sqlx"

var _ PlayLogModel = (*customPlayLogModel)(nil)

type (
	// PlayLogModel is an interface to be customized, add more methods here,
	// and implement the added methods in customPlayLogModel.
	PlayLogModel interface {
		playLogModel
		withSession(session sqlx.Session) PlayLogModel
	}

	customPlayLogModel struct {
		*defaultPlayLogModel
	}
)

// NewPlayLogModel returns a model for the database table.
func NewPlayLogModel(conn sqlx.SqlConn) PlayLogModel {
	return &customPlayLogModel{
		defaultPlayLogModel: newPlayLogModel(conn),
	}
}

func (m *customPlayLogModel) withSession(session sqlx.Session) PlayLogModel {
	return NewPlayLogModel(sqlx.NewSqlConnFromSession(session))
}

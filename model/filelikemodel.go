package model

import "github.com/zeromicro/go-zero/core/stores/sqlx"

var _ FileLikeModel = (*customFileLikeModel)(nil)

type (
	// FileLikeModel is an interface to be customized, add more methods here,
	// and implement the added methods in customFileLikeModel.
	FileLikeModel interface {
		fileLikeModel
		withSession(session sqlx.Session) FileLikeModel
	}

	customFileLikeModel struct {
		*defaultFileLikeModel
	}
)

// NewFileLikeModel returns a model for the database table.
func NewFileLikeModel(conn sqlx.SqlConn) FileLikeModel {
	return &customFileLikeModel{
		defaultFileLikeModel: newFileLikeModel(conn),
	}
}

func (m *customFileLikeModel) withSession(session sqlx.Session) FileLikeModel {
	return NewFileLikeModel(sqlx.NewSqlConnFromSession(session))
}

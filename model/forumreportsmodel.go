package model

import "github.com/zeromicro/go-zero/core/stores/sqlx"

var _ ForumReportsModel = (*customForumReportsModel)(nil)

type (
	// ForumReportsModel is an interface to be customized, add more methods here,
	// and implement the added methods in customForumReportsModel.
	ForumReportsModel interface {
		forumReportsModel
		withSession(session sqlx.Session) ForumReportsModel
	}

	customForumReportsModel struct {
		*defaultForumReportsModel
	}
)

// NewForumReportsModel returns a model for the database table.
func NewForumReportsModel(conn sqlx.SqlConn) ForumReportsModel {
	return &customForumReportsModel{
		defaultForumReportsModel: newForumReportsModel(conn),
	}
}

func (m *customForumReportsModel) withSession(session sqlx.Session) ForumReportsModel {
	return NewForumReportsModel(sqlx.NewSqlConnFromSession(session))
}

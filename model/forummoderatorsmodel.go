package model

import "github.com/zeromicro/go-zero/core/stores/sqlx"

var _ ForumModeratorsModel = (*customForumModeratorsModel)(nil)

type (
	// ForumModeratorsModel is an interface to be customized, add more methods here,
	// and implement the added methods in customForumModeratorsModel.
	ForumModeratorsModel interface {
		forumModeratorsModel
		withSession(session sqlx.Session) ForumModeratorsModel
	}

	customForumModeratorsModel struct {
		*defaultForumModeratorsModel
	}
)

// NewForumModeratorsModel returns a model for the database table.
func NewForumModeratorsModel(conn sqlx.SqlConn) ForumModeratorsModel {
	return &customForumModeratorsModel{
		defaultForumModeratorsModel: newForumModeratorsModel(conn),
	}
}

func (m *customForumModeratorsModel) withSession(session sqlx.Session) ForumModeratorsModel {
	return NewForumModeratorsModel(sqlx.NewSqlConnFromSession(session))
}

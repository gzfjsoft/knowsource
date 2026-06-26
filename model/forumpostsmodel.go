package model

import "github.com/zeromicro/go-zero/core/stores/sqlx"

var _ ForumPostsModel = (*customForumPostsModel)(nil)

type (
	// ForumPostsModel is an interface to be customized, add more methods here,
	// and implement the added methods in customForumPostsModel.
	ForumPostsModel interface {
		forumPostsModel
		withSession(session sqlx.Session) ForumPostsModel
	}

	customForumPostsModel struct {
		*defaultForumPostsModel
	}
)

// NewForumPostsModel returns a model for the database table.
func NewForumPostsModel(conn sqlx.SqlConn) ForumPostsModel {
	return &customForumPostsModel{
		defaultForumPostsModel: newForumPostsModel(conn),
	}
}

func (m *customForumPostsModel) withSession(session sqlx.Session) ForumPostsModel {
	return NewForumPostsModel(sqlx.NewSqlConnFromSession(session))
}

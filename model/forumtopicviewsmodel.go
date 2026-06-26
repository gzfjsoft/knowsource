package model

import "github.com/zeromicro/go-zero/core/stores/sqlx"

var _ ForumTopicViewsModel = (*customForumTopicViewsModel)(nil)

type (
	// ForumTopicViewsModel is an interface to be customized, add more methods here,
	// and implement the added methods in customForumTopicViewsModel.
	ForumTopicViewsModel interface {
		forumTopicViewsModel
		withSession(session sqlx.Session) ForumTopicViewsModel
	}

	customForumTopicViewsModel struct {
		*defaultForumTopicViewsModel
	}
)

// NewForumTopicViewsModel returns a model for the database table.
func NewForumTopicViewsModel(conn sqlx.SqlConn) ForumTopicViewsModel {
	return &customForumTopicViewsModel{
		defaultForumTopicViewsModel: newForumTopicViewsModel(conn),
	}
}

func (m *customForumTopicViewsModel) withSession(session sqlx.Session) ForumTopicViewsModel {
	return NewForumTopicViewsModel(sqlx.NewSqlConnFromSession(session))
}

package model

import "github.com/zeromicro/go-zero/core/stores/sqlx"

var _ ForumCategoriesModel = (*customForumCategoriesModel)(nil)

type (
	// ForumCategoriesModel is an interface to be customized, add more methods here,
	// and implement the added methods in customForumCategoriesModel.
	ForumCategoriesModel interface {
		forumCategoriesModel
		withSession(session sqlx.Session) ForumCategoriesModel
	}

	customForumCategoriesModel struct {
		*defaultForumCategoriesModel
	}
)

// NewForumCategoriesModel returns a model for the database table.
func NewForumCategoriesModel(conn sqlx.SqlConn) ForumCategoriesModel {
	return &customForumCategoriesModel{
		defaultForumCategoriesModel: newForumCategoriesModel(conn),
	}
}

func (m *customForumCategoriesModel) withSession(session sqlx.Session) ForumCategoriesModel {
	return NewForumCategoriesModel(sqlx.NewSqlConnFromSession(session))
}

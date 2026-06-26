package model

import (
	"context"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ ForumForumsModel = (*customForumForumsModel)(nil)

type (
	// ForumForumsModel is an interface to be customized, add more methods here,
	// and implement the added methods in customForumForumsModel.
	ForumForumsModel interface {
		forumForumsModel
		withSession(session sqlx.Session) ForumForumsModel
		CheckForumExists(ctx context.Context, forumId uint64) (bool, error)
	}

	customForumForumsModel struct {
		*defaultForumForumsModel
	}
)

// NewForumForumsModel returns a model for the database table.
func NewForumForumsModel(conn sqlx.SqlConn) ForumForumsModel {
	return &customForumForumsModel{
		defaultForumForumsModel: newForumForumsModel(conn),
	}
}

func (m *customForumForumsModel) withSession(session sqlx.Session) ForumForumsModel {
	return NewForumForumsModel(sqlx.NewSqlConnFromSession(session))
}

// CheckForumExists checks if a forum exists and is active
func (m *customForumForumsModel) CheckForumExists(ctx context.Context, forumId uint64) (bool, error) {
	var count int64
	query := "SELECT COUNT(*) FROM forum_forums WHERE forum_id = ? AND is_active = 1"
	err := m.conn.QueryRowCtx(ctx, &count, query, forumId)
	return count > 0, err
}

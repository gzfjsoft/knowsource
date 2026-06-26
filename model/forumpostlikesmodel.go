package model

import (
	"context"
	"fmt"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ ForumPostLikesModel = (*customForumPostLikesModel)(nil)

type (
	// ForumPostLikesModel is an interface to be customized, add more methods here,
	// and implement the added methods in customForumPostLikesModel.
	ForumPostLikesModel interface {
		forumPostLikesModel
		withSession(session sqlx.Session) ForumPostLikesModel
		DeleteByUserIdAndPostId(ctx context.Context, userId, postId uint64) error
	}

	customForumPostLikesModel struct {
		*defaultForumPostLikesModel
	}
)

// NewForumPostLikesModel returns a model for the database table.
func NewForumPostLikesModel(conn sqlx.SqlConn) ForumPostLikesModel {
	return &customForumPostLikesModel{
		defaultForumPostLikesModel: newForumPostLikesModel(conn),
	}
}

func (m *customForumPostLikesModel) withSession(session sqlx.Session) ForumPostLikesModel {
	return NewForumPostLikesModel(sqlx.NewSqlConnFromSession(session))
}

func (m *customForumPostLikesModel) DeleteByUserIdAndPostId(ctx context.Context, userId, postId uint64) error {
	query := fmt.Sprintf("delete from %s where `user_id` = ? and `post_id` = ?", m.table)
	_, err := m.conn.ExecCtx(ctx, query, userId, postId)
	return err
}

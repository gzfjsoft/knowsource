package model

import (
	"context"
	"time"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ ForumUserStatsModel = (*customForumUserStatsModel)(nil)

type (
	// ForumUserStatsModel is an interface to be customized, add more methods here,
	// and implement the added methods in customForumUserStatsModel.
	ForumUserStatsModel interface {
		forumUserStatsModel
		withSession(session sqlx.Session) ForumUserStatsModel
		UpdateUserStats(ctx context.Context, userId uint64, topicCount, postCount int) error
	}

	customForumUserStatsModel struct {
		*defaultForumUserStatsModel
	}
)

// NewForumUserStatsModel returns a model for the database table.
func NewForumUserStatsModel(conn sqlx.SqlConn) ForumUserStatsModel {
	return &customForumUserStatsModel{
		defaultForumUserStatsModel: newForumUserStatsModel(conn),
	}
}

func (m *customForumUserStatsModel) withSession(session sqlx.Session) ForumUserStatsModel {
	return NewForumUserStatsModel(sqlx.NewSqlConnFromSession(session))
}

// UpdateUserStats updates user statistics
func (m *customForumUserStatsModel) UpdateUserStats(ctx context.Context, userId uint64, topicCount, postCount int) error {
	now := time.Now()
	query := `INSERT INTO forum_user_stats (user_id, topic_count, post_count, like_received, like_given, last_activity_at, created_at, updated_at) 
			  VALUES (?, ?, ?, 0, 0, ?, NOW(), NOW()) 
			  ON DUPLICATE KEY UPDATE topic_count = topic_count + ?, post_count = post_count + ?, last_activity_at = ?, updated_at = NOW()`

	_, err := m.conn.ExecCtx(ctx, query, userId, topicCount, postCount, now, topicCount, postCount, now)
	return err
}

package model

import (
	"context"
	"fmt"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ ForumTopicSubscriptionsModel = (*customForumTopicSubscriptionsModel)(nil)

type (
	// ForumTopicSubscriptionsModel is an interface to be customized, add more methods here,
	// and implement the added methods in customForumTopicSubscriptionsModel.
	ForumTopicSubscriptionsModel interface {
		forumTopicSubscriptionsModel
		withSession(session sqlx.Session) ForumTopicSubscriptionsModel
		FindByUserId(ctx context.Context, userId uint64, offset, limit int64) ([]*ForumTopicSubscriptions, error)
		CountByUserId(ctx context.Context, userId uint64) (int64, error)
		DeleteByUserIdAndTopicId(ctx context.Context, userId, topicId uint64) error
	}

	customForumTopicSubscriptionsModel struct {
		*defaultForumTopicSubscriptionsModel
	}
)

// NewForumTopicSubscriptionsModel returns a model for the database table.
func NewForumTopicSubscriptionsModel(conn sqlx.SqlConn) ForumTopicSubscriptionsModel {
	return &customForumTopicSubscriptionsModel{
		defaultForumTopicSubscriptionsModel: newForumTopicSubscriptionsModel(conn),
	}
}

func (m *customForumTopicSubscriptionsModel) withSession(session sqlx.Session) ForumTopicSubscriptionsModel {
	return NewForumTopicSubscriptionsModel(sqlx.NewSqlConnFromSession(session))
}

func (m *customForumTopicSubscriptionsModel) FindByUserId(ctx context.Context, userId uint64, offset, limit int64) ([]*ForumTopicSubscriptions, error) {
	query := fmt.Sprintf("select %s from %s where `user_id` = ? order by `created_at` desc limit ?, ?", forumTopicSubscriptionsRows, m.table)
	var resp []*ForumTopicSubscriptions
	err := m.conn.QueryRowsCtx(ctx, &resp, query, userId, offset, limit)
	return resp, err
}

func (m *customForumTopicSubscriptionsModel) CountByUserId(ctx context.Context, userId uint64) (int64, error) {
	query := fmt.Sprintf("select count(*) from %s where `user_id` = ?", m.table)
	var count int64
	err := m.conn.QueryRowCtx(ctx, &count, query, userId)
	return count, err
}

func (m *customForumTopicSubscriptionsModel) DeleteByUserIdAndTopicId(ctx context.Context, userId, topicId uint64) error {
	query := fmt.Sprintf("delete from %s where `user_id` = ? and `topic_id` = ?", m.table)
	_, err := m.conn.ExecCtx(ctx, query, userId, topicId)
	return err
}

package model

import (
	"context"
	"fmt"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ ForumNotificationsModel = (*customForumNotificationsModel)(nil)

type (
	// ForumNotificationsModel is an interface to be customized, add more methods here,
	// and implement the added methods in customForumNotificationsModel.
	ForumNotificationsModel interface {
		forumNotificationsModel
		withSession(session sqlx.Session) ForumNotificationsModel
		FindByUserId(ctx context.Context, userId uint64, isRead int32, offset, limit int64) ([]*ForumNotifications, error)
		CountByUserId(ctx context.Context, userId uint64, isRead int32) (int64, error)
		MarkAllAsReadByUserId(ctx context.Context, userId uint64) error
	}

	customForumNotificationsModel struct {
		*defaultForumNotificationsModel
	}
)

// NewForumNotificationsModel returns a model for the database table.
func NewForumNotificationsModel(conn sqlx.SqlConn) ForumNotificationsModel {
	return &customForumNotificationsModel{
		defaultForumNotificationsModel: newForumNotificationsModel(conn),
	}
}

func (m *customForumNotificationsModel) withSession(session sqlx.Session) ForumNotificationsModel {
	return NewForumNotificationsModel(sqlx.NewSqlConnFromSession(session))
}

func (m *customForumNotificationsModel) FindByUserId(ctx context.Context, userId uint64, isRead int32, offset, limit int64) ([]*ForumNotifications, error) {
	var query string
	var args []interface{}

	if isRead == 0 {
		// Get all notifications
		query = fmt.Sprintf("select %s from %s where `user_id` = ? order by `created_at` desc limit ?, ?", forumNotificationsRows, m.table)
		args = []interface{}{userId, offset, limit}
	} else {
		// Get notifications by read status
		query = fmt.Sprintf("select %s from %s where `user_id` = ? and `is_read` = ? order by `created_at` desc limit ?, ?", forumNotificationsRows, m.table)
		args = []interface{}{userId, isRead, offset, limit}
	}

	var resp []*ForumNotifications
	err := m.conn.QueryRowsCtx(ctx, &resp, query, args...)
	return resp, err
}

func (m *customForumNotificationsModel) CountByUserId(ctx context.Context, userId uint64, isRead int32) (int64, error) {
	var query string
	var args []interface{}

	if isRead == 0 {
		// Count all notifications
		query = fmt.Sprintf("select count(*) from %s where `user_id` = ?", m.table)
		args = []interface{}{userId}
	} else {
		// Count notifications by read status
		query = fmt.Sprintf("select count(*) from %s where `user_id` = ? and `is_read` = ?", m.table)
		args = []interface{}{userId, isRead}
	}

	var count int64
	err := m.conn.QueryRowCtx(ctx, &count, query, args...)
	return count, err
}

func (m *customForumNotificationsModel) MarkAllAsReadByUserId(ctx context.Context, userId uint64) error {
	query := fmt.Sprintf("update %s set `is_read` = 1 where `user_id` = ? and `is_read` = 0", m.table)
	_, err := m.conn.ExecCtx(ctx, query, userId)
	return err
}

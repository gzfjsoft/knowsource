package model

import (
	"context"
	"fmt"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ ForumBansModel = (*customForumBansModel)(nil)

type (
	// ForumBansModel is an interface to be customized, add more methods here,
	// and implement the added methods in customForumBansModel.
	ForumBansModel interface {
		forumBansModel
		withSession(session sqlx.Session) ForumBansModel
		FindByForumId(ctx context.Context, forumId uint64, offset, limit int64) ([]*ForumBans, error)
		FindByUserId(ctx context.Context, userId uint64) ([]*ForumBans, error)
		CountByForumId(ctx context.Context, forumId uint64) (int64, error)
		FindActiveBanByUserAndForum(ctx context.Context, userId, forumId uint64) (*ForumBans, error)
	}

	customForumBansModel struct {
		*defaultForumBansModel
	}
)

// NewForumBansModel returns a model for the database table.
func NewForumBansModel(conn sqlx.SqlConn) ForumBansModel {
	return &customForumBansModel{
		defaultForumBansModel: newForumBansModel(conn),
	}
}

func (m *customForumBansModel) withSession(session sqlx.Session) ForumBansModel {
	return NewForumBansModel(sqlx.NewSqlConnFromSession(session))
}

func (m *customForumBansModel) FindByForumId(ctx context.Context, forumId uint64, offset, limit int64) ([]*ForumBans, error) {
	query := fmt.Sprintf("select %s from %s where `forum_id` = ? order by `created_at` desc limit ? offset ?", forumBansRows, m.table)
	var resp []*ForumBans
	err := m.conn.QueryRowsCtx(ctx, &resp, query, forumId, limit, offset)
	return resp, err
}

func (m *customForumBansModel) FindByUserId(ctx context.Context, userId uint64) ([]*ForumBans, error) {
	query := fmt.Sprintf("select %s from %s where `user_id` = ? order by `created_at` desc", forumBansRows, m.table)
	var resp []*ForumBans
	err := m.conn.QueryRowsCtx(ctx, &resp, query, userId)
	return resp, err
}

func (m *customForumBansModel) CountByForumId(ctx context.Context, forumId uint64) (int64, error) {
	query := fmt.Sprintf("select count(*) from %s where `forum_id` = ?", m.table)
	var count int64
	err := m.conn.QueryRowCtx(ctx, &count, query, forumId)
	return count, err
}

func (m *customForumBansModel) FindActiveBanByUserAndForum(ctx context.Context, userId, forumId uint64) (*ForumBans, error) {
	query := fmt.Sprintf("select %s from %s where `user_id` = ? and `forum_id` = ? and (`is_permanent` = 1 or `expires_at` > now()) limit 1", forumBansRows, m.table)
	var resp ForumBans
	err := m.conn.QueryRowCtx(ctx, &resp, query, userId, forumId)
	switch err {
	case nil:
		return &resp, nil
	case sqlx.ErrNotFound:
		return nil, ErrNotFound
	default:
		return nil, err
	}
}

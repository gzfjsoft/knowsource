package model

import (
	"context"
	"fmt"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ ForumAttachmentsModel = (*customForumAttachmentsModel)(nil)

type (
	// ForumAttachmentsModel is an interface to be customized, add more methods here,
	// and implement the added methods in customForumAttachmentsModel.
	ForumAttachmentsModel interface {
		forumAttachmentsModel
		withSession(session sqlx.Session) ForumAttachmentsModel
		FindByFilters(ctx context.Context, topicId, postId uint64, offset, limit int64) ([]*ForumAttachments, error)
		CountByFilters(ctx context.Context, topicId, postId uint64) (int64, error)
	}

	customForumAttachmentsModel struct {
		*defaultForumAttachmentsModel
	}
)

// NewForumAttachmentsModel returns a model for the database table.
func NewForumAttachmentsModel(conn sqlx.SqlConn) ForumAttachmentsModel {
	return &customForumAttachmentsModel{
		defaultForumAttachmentsModel: newForumAttachmentsModel(conn),
	}
}

func (m *customForumAttachmentsModel) withSession(session sqlx.Session) ForumAttachmentsModel {
	return NewForumAttachmentsModel(sqlx.NewSqlConnFromSession(session))
}

func (m *customForumAttachmentsModel) FindByFilters(ctx context.Context, topicId, postId uint64, offset, limit int64) ([]*ForumAttachments, error) {
	var query string
	var args []interface{}

	if topicId > 0 && postId > 0 {
		// Filter by both topic and post
		query = fmt.Sprintf("select %s from %s where `topic_id` = ? and `post_id` = ? order by `created_at` desc limit ?, ?", forumAttachmentsRows, m.table)
		args = []interface{}{topicId, postId, offset, limit}
	} else if topicId > 0 {
		// Filter by topic only
		query = fmt.Sprintf("select %s from %s where `topic_id` = ? order by `created_at` desc limit ?, ?", forumAttachmentsRows, m.table)
		args = []interface{}{topicId, offset, limit}
	} else if postId > 0 {
		// Filter by post only
		query = fmt.Sprintf("select %s from %s where `post_id` = ? order by `created_at` desc limit ?, ?", forumAttachmentsRows, m.table)
		args = []interface{}{postId, offset, limit}
	} else {
		// No filters, get all
		query = fmt.Sprintf("select %s from %s order by `created_at` desc limit ?, ?", forumAttachmentsRows, m.table)
		args = []interface{}{offset, limit}
	}

	var resp []*ForumAttachments
	err := m.conn.QueryRowsCtx(ctx, &resp, query, args...)
	return resp, err
}

func (m *customForumAttachmentsModel) CountByFilters(ctx context.Context, topicId, postId uint64) (int64, error) {
	var query string
	var args []interface{}

	if topicId > 0 && postId > 0 {
		// Count by both topic and post
		query = fmt.Sprintf("select count(*) from %s where `topic_id` = ? and `post_id` = ?", m.table)
		args = []interface{}{topicId, postId}
	} else if topicId > 0 {
		// Count by topic only
		query = fmt.Sprintf("select count(*) from %s where `topic_id` = ?", m.table)
		args = []interface{}{topicId}
	} else if postId > 0 {
		// Count by post only
		query = fmt.Sprintf("select count(*) from %s where `post_id` = ?", m.table)
		args = []interface{}{postId}
	} else {
		// Count all
		query = fmt.Sprintf("select count(*) from %s", m.table)
		args = []interface{}{}
	}

	var count int64
	err := m.conn.QueryRowCtx(ctx, &count, query, args...)
	return count, err
}

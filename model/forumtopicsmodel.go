package model

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ ForumTopicsModel = (*customForumTopicsModel)(nil)

type (
	// ForumTopicsModel is an interface to be customized, add more methods here,
	// and implement the added methods in customForumTopicsModel.
	ForumTopicsModel interface {
		forumTopicsModel
		withSession(session sqlx.Session) ForumTopicsModel
		CreateTopicWithPost(ctx context.Context, forumId, userId uint64, title, content, slug string, isSticky, isLocked, isAnnouncement int32, postsModel ForumPostsModel) (int64, error)
		// New methods for list, get, and search functionality
		ListTopics(ctx context.Context, forumId, userId uint64, page, pageSize int64) ([]*ForumTopics, int64, error)
		SearchTopics(ctx context.Context, query string, forumId uint64, page, pageSize int64) ([]*ForumTopics, int64, error)
		GetTopicWithViewCount(ctx context.Context, topicId uint64) (*ForumTopics, error)
		IncrementViewCount(ctx context.Context, topicId uint64) error
		RecordTopicView(ctx context.Context, topicId, userId uint64, ipAddress, userAgent string) error
	}

	customForumTopicsModel struct {
		*defaultForumTopicsModel
	}
)

// NewForumTopicsModel returns a model for the database table.
func NewForumTopicsModel(conn sqlx.SqlConn) ForumTopicsModel {
	return &customForumTopicsModel{
		defaultForumTopicsModel: newForumTopicsModel(conn),
	}
}

func (m *customForumTopicsModel) withSession(session sqlx.Session) ForumTopicsModel {
	return NewForumTopicsModel(sqlx.NewSqlConnFromSession(session))
}

// CreateTopicWithPost creates a topic and its first post in a transaction
func (m *customForumTopicsModel) CreateTopicWithPost(ctx context.Context, forumId, userId uint64, title, content, slug string, isSticky, isLocked, isAnnouncement int32, postsModel ForumPostsModel) (int64, error) {
	var topicId int64

	err := m.conn.Transact(func(session sqlx.Session) error {
		// Create topic
		now := time.Now()
		topic := &ForumTopics{
			ForumId:         forumId,
			UserId:          userId,
			Title:           title,
			Content:         sql.NullString{String: content, Valid: true},
			Slug:            slug,
			IsSticky:        uint64(isSticky),
			IsLocked:        uint64(isLocked),
			IsAnnouncement:  uint64(isAnnouncement),
			ViewCount:       0,
			ReplyCount:      0,
			LastReplyUserId: 0,
			LastReplyAt:     sql.NullTime{Valid: false},
			CreatedAt:       now,
			UpdatedAt:       now,
		}

		result, err := m.withSession(session).Insert(ctx, topic)
		if err != nil {
			return err
		}

		topicId, err = result.LastInsertId()
		if err != nil {
			return err
		}

		// Create first post
		post := &ForumPosts{
			TopicId:     uint64(topicId),
			UserId:      userId,
			ParentId:    0,
			Content:     sql.NullString{String: content, Valid: true},
			IsFirstPost: 1,
			CreatedAt:   now,
			UpdatedAt:   now,
		}

		_, err = postsModel.withSession(session).Insert(ctx, post)
		return err
	})

	return topicId, err
}

// ListTopics retrieves a list of topics with filtering and pagination
func (m *customForumTopicsModel) ListTopics(ctx context.Context, forumId, userId uint64, page, pageSize int64) ([]*ForumTopics, int64, error) {
	// Build query conditions
	var conditions []string
	var args []interface{}

	if forumId > 0 {
		conditions = append(conditions, "forum_id = ?")
		args = append(args, forumId)
	}

	if userId > 0 {
		conditions = append(conditions, "user_id = ?")
		args = append(args, userId)
	}

	// Build WHERE clause
	whereClause := ""
	if len(conditions) > 0 {
		whereClause = "WHERE " + strings.Join(conditions, " AND ")
	}

	// Build query with ordering
	query := fmt.Sprintf("SELECT %s FROM %s %s ORDER BY is_sticky DESC, last_reply_at DESC, created_at DESC LIMIT ? OFFSET ?",
		forumTopicsRows, m.table, whereClause)

	// Add pagination parameters
	args = append(args, pageSize, (page-1)*pageSize)

	// Execute query
	var topics []*ForumTopics
	if err := m.conn.QueryRowsCtx(ctx, &topics, query, args...); err != nil {
		return nil, 0, err
	}

	// Get total count
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM %s %s", m.table, whereClause)
	var total int64
	if err := m.conn.QueryRowCtx(ctx, &total, countQuery, args[:len(args)-2]...); err != nil {
		return nil, 0, err
	}

	return topics, total, nil
}

// SearchTopics searches topics by title and content with filtering and pagination
func (m *customForumTopicsModel) SearchTopics(ctx context.Context, query string, forumId uint64, page, pageSize int64) ([]*ForumTopics, int64, error) {
	// Build query conditions
	var conditions []string
	var args []interface{}

	// Add search keywords
	searchTerm := "%" + query + "%"
	conditions = append(conditions, "(title LIKE ? OR content LIKE ?)")
	args = append(args, searchTerm, searchTerm)

	// Add forum filter
	if forumId > 0 {
		conditions = append(conditions, "forum_id = ?")
		args = append(args, forumId)
	}

	// Build WHERE clause
	whereClause := "WHERE " + strings.Join(conditions, " AND ")

	// Build query with ordering
	searchQuery := fmt.Sprintf("SELECT %s FROM %s %s ORDER BY is_sticky DESC, last_reply_at DESC, created_at DESC LIMIT ? OFFSET ?",
		forumTopicsRows, m.table, whereClause)

	// Add pagination parameters
	args = append(args, pageSize, (page-1)*pageSize)

	// Execute query
	var topics []*ForumTopics
	if err := m.conn.QueryRowsCtx(ctx, &topics, searchQuery, args...); err != nil {
		return nil, 0, err
	}

	// Get total count
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM %s %s", m.table, whereClause)
	var total int64
	if err := m.conn.QueryRowCtx(ctx, &total, countQuery, args[:len(args)-2]...); err != nil {
		return nil, 0, err
	}

	return topics, total, nil
}

// GetTopicWithViewCount retrieves a topic and increments its view count
func (m *customForumTopicsModel) GetTopicWithViewCount(ctx context.Context, topicId uint64) (*ForumTopics, error) {
	// Get topic
	topic, err := m.FindOne(ctx, topicId)
	if err != nil {
		return nil, err
	}

	// Increment view count
	err = m.IncrementViewCount(ctx, topicId)
	if err != nil {
		return nil, err
	}

	return topic, nil
}

// IncrementViewCount increments the view count of a topic
func (m *customForumTopicsModel) IncrementViewCount(ctx context.Context, topicId uint64) error {
	query := fmt.Sprintf("UPDATE %s SET view_count = view_count + 1 WHERE topic_id = ?", m.table)
	_, err := m.conn.ExecCtx(ctx, query, topicId)
	return err
}

// RecordTopicView records a topic view for a user
func (m *customForumTopicsModel) RecordTopicView(ctx context.Context, topicId, userId uint64, ipAddress, userAgent string) error {
	query := `INSERT INTO forum_topic_views (topic_id, user_id, ip_address, user_agent, viewed_at) 
			  VALUES (?, ?, ?, ?, ?)`
	_, err := m.conn.ExecCtx(ctx, query, topicId, userId, ipAddress, userAgent, time.Now())
	return err
}

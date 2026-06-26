package model

import (
	"context"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ AiSessionsModel = (*customAiSessionsModel)(nil)

type (
	// AiSessionsModel is an interface to be customized, add more methods here,
	// and implement the added methods in customAiSessionsModel.
	AiSessionsModel interface {
		aiSessionsModel
		withSession(session sqlx.Session) AiSessionsModel
		FindByUserId(ctx context.Context, clientId string, userId int64) ([]*AiSessions, error)
		FindByUserIdWithLastMessage(ctx context.Context, clientId string, userId int64) ([]*AiSessionsWithLastMessage, error)
		FindByEmpCodeWithLastMessage(ctx context.Context, clientId string, empCode string) ([]*AiSessionsWithLastMessage, error)
	}

	customAiSessionsModel struct {
		*defaultAiSessionsModel
	}

	// AiSessionsWithLastMessage 包含会话信息和最后一条消息的查询结果
	AiSessionsWithLastMessage struct {
		AiSessions
		Title     string `db:"title"`      // 第一条消息作为title
		LastQuery string `db:"last_query"` // 最后一条用户消息
		LastReply string `db:"last_reply"` // 最后一条AI回复
	}
)

// NewAiSessionsModel returns a model for the database table.
func NewAiSessionsModel(conn sqlx.SqlConn) AiSessionsModel {
	return &customAiSessionsModel{
		defaultAiSessionsModel: newAiSessionsModel(conn),
	}
}

func (m *customAiSessionsModel) withSession(session sqlx.Session) AiSessionsModel {
	return NewAiSessionsModel(sqlx.NewSqlConnFromSession(session))
}

// FindByUserId 根据租户和用户ID查询会话列表
func (m *customAiSessionsModel) FindByUserId(ctx context.Context, clientId string, userId int64) ([]*AiSessions, error) {
	query := `SELECT * FROM ai_sessions WHERE client_id = ? AND user_id = ? AND session_status != 'deleted' ORDER BY last_message_time DESC, created_at DESC`
	var sessions []*AiSessions
	err := m.conn.QueryRowsCtx(ctx, &sessions, query, clientId, userId)
	return sessions, err
}

// FindByUserIdWithLastMessage 根据租户和用户ID查询会话列表，包含最后一条消息
func (m *customAiSessionsModel) FindByUserIdWithLastMessage(ctx context.Context, clientId string, userId int64) ([]*AiSessionsWithLastMessage, error) {
	query := `
		SELECT 
			s.*,
			COALESCE(first_message.content, '') as title,
			COALESCE(last_user.content, '') as last_query,
			COALESCE(last_assistant.content, '') as last_reply
		FROM ai_sessions s
		LEFT JOIN (
			SELECT session_id, content,
				   ROW_NUMBER() OVER (PARTITION BY session_id ORDER BY created_at ASC) as rn
			FROM ai_messages 
			WHERE role = 'user'
		) first_message ON s.session_id = first_message.session_id AND first_message.rn = 1
		LEFT JOIN (
			SELECT session_id, content,
				   ROW_NUMBER() OVER (PARTITION BY session_id ORDER BY created_at DESC) as rn
			FROM ai_messages 
			WHERE role = 'user'
		) last_user ON s.session_id = last_user.session_id AND last_user.rn = 1
		LEFT JOIN (
			SELECT session_id, content,
				   ROW_NUMBER() OVER (PARTITION BY session_id ORDER BY created_at DESC) as rn
			FROM ai_messages 
			WHERE role = 'assistant'
		) last_assistant ON s.session_id = last_assistant.session_id AND last_assistant.rn = 1
		WHERE s.client_id = ? AND s.user_id = ? AND s.session_status != 'deleted'
		ORDER BY s.last_message_time DESC, s.created_at DESC
	`
	var sessions []*AiSessionsWithLastMessage
	err := m.conn.QueryRowsCtx(ctx, &sessions, query, clientId, userId)
	return sessions, err
}

// FindByEmpCodeWithLastMessage 根据租户和员工编码查询会话列表，包含最后一条消息
func (m *customAiSessionsModel) FindByEmpCodeWithLastMessage(ctx context.Context, clientId string, empCode string) ([]*AiSessionsWithLastMessage, error) {
	query := `
		SELECT 
			s.*,
			COALESCE(first_message.content, '') as title,
			COALESCE(last_user.content, '') as last_query,
			COALESCE(last_assistant.content, '') as last_reply
		FROM ai_sessions s
		LEFT JOIN (
			SELECT session_id, content,
				   ROW_NUMBER() OVER (PARTITION BY session_id ORDER BY created_at ASC) as rn
			FROM ai_messages 
			WHERE role = 'user'
		) first_message ON s.session_id = first_message.session_id AND first_message.rn = 1
		LEFT JOIN (
			SELECT session_id, content,
				   ROW_NUMBER() OVER (PARTITION BY session_id ORDER BY created_at DESC) as rn
			FROM ai_messages 
			WHERE role = 'user'
		) last_user ON s.session_id = last_user.session_id AND last_user.rn = 1
		LEFT JOIN (
			SELECT session_id, content,
				   ROW_NUMBER() OVER (PARTITION BY session_id ORDER BY created_at DESC) as rn
			FROM ai_messages 
			WHERE role = 'assistant'
		) last_assistant ON s.session_id = last_assistant.session_id AND last_assistant.rn = 1
		WHERE s.client_id = ? AND s.emp_code = ? AND s.session_status != 'deleted'
		ORDER BY s.last_message_time DESC, s.created_at DESC
	`
	var sessions []*AiSessionsWithLastMessage
	err := m.conn.QueryRowsCtx(ctx, &sessions, query, clientId, empCode)
	return sessions, err
}

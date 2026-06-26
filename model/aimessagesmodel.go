package model

import (
	"context"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ AiMessagesModel = (*customAiMessagesModel)(nil)

type (
	// AiMessagesModel is an interface to be customized, add more methods here,
	// and implement the added methods in customAiMessagesModel.
	AiMessagesModel interface {
		aiMessagesModel
		withSession(session sqlx.Session) AiMessagesModel
		FindBySessionId(ctx context.Context, sessionId uint64) ([]*AiMessages, error)
		FindBySessionUuid(ctx context.Context, sessionUuid string) ([]*AiMessages, error)
	}

	customAiMessagesModel struct {
		*defaultAiMessagesModel
	}
)

// NewAiMessagesModel returns a model for the database table.
func NewAiMessagesModel(conn sqlx.SqlConn) AiMessagesModel {
	return &customAiMessagesModel{
		defaultAiMessagesModel: newAiMessagesModel(conn),
	}
}

func (m *customAiMessagesModel) withSession(session sqlx.Session) AiMessagesModel {
	return NewAiMessagesModel(sqlx.NewSqlConnFromSession(session))
}

// FindBySessionId 根据会话ID查询消息列表
func (m *customAiMessagesModel) FindBySessionId(ctx context.Context, sessionId uint64) ([]*AiMessages, error) {
	query := `SELECT * FROM ai_messages WHERE session_id = ? ORDER BY created_at ASC`
	var messages []*AiMessages
	err := m.conn.QueryRowsCtx(ctx, &messages, query, sessionId)
	return messages, err
}

// FindBySessionUuid 根据会话UUID查询消息列表
func (m *customAiMessagesModel) FindBySessionUuid(ctx context.Context, sessionUuid string) ([]*AiMessages, error) {
	query := `
		SELECT m.* 
		FROM ai_messages m
		JOIN ai_sessions s ON m.session_id = s.session_id
		WHERE s.session_uuid = ? 
		ORDER BY m.created_at ASC
	`
	var messages []*AiMessages
	err := m.conn.QueryRowsCtx(ctx, &messages, query, sessionUuid)
	return messages, err
}

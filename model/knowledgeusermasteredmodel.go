package model

import (
	"context"
	"fmt"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ KnowledgeUserMasteredModel = (*customKnowledgeUserMasteredModel)(nil)

type (
	// KnowledgeUserMasteredModel is an interface to be customized, add more methods here,
	// and implement the added methods in customKnowledgeUserMasteredModel.
	KnowledgeUserMasteredModel interface {
		knowledgeUserMasteredModel
		withSession(session sqlx.Session) KnowledgeUserMasteredModel
		FindOneByUserIdAndKpId(ctx context.Context, userId, kpId uint64) (*KnowledgeUserMastered, error)
	}

	customKnowledgeUserMasteredModel struct {
		*defaultKnowledgeUserMasteredModel
	}
)

// NewKnowledgeUserMasteredModel returns a model for the database table.
func NewKnowledgeUserMasteredModel(conn sqlx.SqlConn) KnowledgeUserMasteredModel {
	return &customKnowledgeUserMasteredModel{
		defaultKnowledgeUserMasteredModel: newKnowledgeUserMasteredModel(conn),
	}
}

func (m *customKnowledgeUserMasteredModel) withSession(session sqlx.Session) KnowledgeUserMasteredModel {
	return NewKnowledgeUserMasteredModel(sqlx.NewSqlConnFromSession(session))
}

func (m *customKnowledgeUserMasteredModel) FindOneByUserIdAndKpId(ctx context.Context, userId, kpId uint64) (*KnowledgeUserMastered, error) {
	query := fmt.Sprintf("select %s from %s where `user_id` = ? and `kp_id` = ? limit 1", knowledgeUserMasteredRows, m.table)
	var resp KnowledgeUserMastered
	err := m.conn.QueryRowCtx(ctx, &resp, query, userId, kpId)
	switch err {
	case nil:
		return &resp, nil
	case sqlx.ErrNotFound:
		return nil, ErrNotFound
	default:
		return nil, err
	}
}

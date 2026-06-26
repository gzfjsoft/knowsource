package model

import (
	"context"
	"fmt"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ KnowledgeResourceModel = (*customKnowledgeResourceModel)(nil)

type (
	// KnowledgeResourceModel is an interface to be customized, add more methods here,
	// and implement the added methods in customKnowledgeResourceModel.
	KnowledgeResourceModel interface {
		knowledgeResourceModel
		withSession(session sqlx.Session) KnowledgeResourceModel
		Count(ctx context.Context, keyword string) (int64, error)
		CountEx(ctx context.Context, keyword string, kpId uint64) (int64, error)
		FindAll(ctx context.Context, page, pageSize int, keyword string) ([]*KnowledgeResource, error)
		FindAllEx(ctx context.Context, page, pageSize int, keyword string, kpId uint64) ([]*KnowledgeResource, error)
	}

	customKnowledgeResourceModel struct {
		*defaultKnowledgeResourceModel
	}
)

// NewKnowledgeResourceModel returns a model for the database table.
func NewKnowledgeResourceModel(conn sqlx.SqlConn) KnowledgeResourceModel {
	return &customKnowledgeResourceModel{
		defaultKnowledgeResourceModel: newKnowledgeResourceModel(conn),
	}
}

func (m *customKnowledgeResourceModel) withSession(session sqlx.Session) KnowledgeResourceModel {
	return NewKnowledgeResourceModel(sqlx.NewSqlConnFromSession(session))
}

func (m *customKnowledgeResourceModel) Count(ctx context.Context, keyword string) (int64, error) {
	var count int64
	query := fmt.Sprintf("select count(*) from %s", m.table)
	args := []interface{}{}

	if keyword != "" {
		query += " where content like ? or type like ?"
		args = append(args, "%"+keyword+"%", "%"+keyword+"%")
	}

	err := m.conn.QueryRowCtx(ctx, &count, query, args...)
	return count, err
}

func (m *customKnowledgeResourceModel) FindAll(ctx context.Context, page, pageSize int, keyword string) ([]*KnowledgeResource, error) {
	query := fmt.Sprintf("select %s from %s", knowledgeResourceRows, m.table)
	args := []interface{}{}

	if keyword != "" {
		query += " where content like ? or type like ?"
		args = append(args, "%"+keyword+"%", "%"+keyword+"%")
	}

	query += " limit ? offset ?"
	args = append(args, pageSize, (page-1)*pageSize)

	var resp []*KnowledgeResource
	err := m.conn.QueryRowsCtx(ctx, &resp, query, args...)
	return resp, err
}

func (m *customKnowledgeResourceModel) CountEx(ctx context.Context, keyword string, kpId uint64) (int64, error) {

	if kpId == 0 {
		return m.Count(ctx, keyword)
	}

	var count int64
	query := fmt.Sprintf("select count(*) from %s where 1=1", m.table)
	args := []interface{}{}

	if kpId != 0 {
		query += " and kr_id in (select kr_id from kp_kr where kp_id = ?)"
		args = append(args, kpId)
	}

	if keyword != "" {

		query += " and ( content like ? or type like ?)"
		args = append(args, "%"+keyword+"%", "%"+keyword+"%")
	}

	err := m.conn.QueryRowCtx(ctx, &count, query, args...)
	return count, err
}

func (m *customKnowledgeResourceModel) FindAllEx(ctx context.Context, page, pageSize int, keyword string, kpId uint64) ([]*KnowledgeResource, error) {
	if kpId == 0 {
		return m.FindAll(ctx, page, pageSize, keyword)
	}

	query := fmt.Sprintf("select %s from %s where 1=1", knowledgeResourceRows, m.table)
	args := []interface{}{}

	if kpId != 0 {
		query += " and kr_id in (select kr_id from kp_kr where kp_id = ?)"
		args = append(args, kpId)
	}

	if keyword != "" {
		query += " and ( content like ? or type like ?)"
		args = append(args, "%"+keyword+"%", "%"+keyword+"%")
	}

	query += " limit ? offset ?"
	args = append(args, pageSize, (page-1)*pageSize)

	var resp []*KnowledgeResource
	err := m.conn.QueryRowsCtx(ctx, &resp, query, args...)
	return resp, err
}

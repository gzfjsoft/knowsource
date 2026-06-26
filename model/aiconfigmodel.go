package model

import (
	"context"
	"fmt"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ AiConfigModel = (*customAiConfigModel)(nil)

type (
	// AiConfigModel is an interface to be customized, add more methods here,
	// and implement the added methods in customAiConfigModel.
	AiConfigModel interface {
		aiConfigModel
		withSession(session sqlx.Session) AiConfigModel
		FindByNameAndCode(ctx context.Context, clientId string, name string, documentCode string) (*AiConfig, error)
		FindListWithDocumentCode(ctx context.Context, clientId string, name string, documentCode string, page, pageSize uint64) ([]*AiConfig, int64, error)
		FindListOrderByDocumentCode(ctx context.Context, clientId string, name string, page, pageSize uint64) ([]*AiConfig, int64, error)
	}

	customAiConfigModel struct {
		*defaultAiConfigModel
	}
)

// NewAiConfigModel returns a model for the database table.
func NewAiConfigModel(conn sqlx.SqlConn) AiConfigModel {
	return &customAiConfigModel{
		defaultAiConfigModel: newAiConfigModel(conn),
	}
}

func (m *customAiConfigModel) withSession(session sqlx.Session) AiConfigModel {
	return NewAiConfigModel(sqlx.NewSqlConnFromSession(session))
}

func (m *customAiConfigModel) FindListOrderByDocumentCode(ctx context.Context, clientId string, name string, page, pageSize uint64) ([]*AiConfig, int64, error) {
	whereConditions := []string{"client_id = ?"}
	args := []interface{}{clientId}
	if name != "" {
		keywordLike := "%" + name + "%"
		whereConditions = append(whereConditions, "name like ?")
		args = append(args, keywordLike)
	}
	whereCondition := " where " + whereConditions[0]
	for i := 1; i < len(whereConditions); i++ {
		whereCondition = whereCondition + " and " + whereConditions[i]
	}

	// 查询总数
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM %s %s", m.table, whereCondition)
	var count int64
	err := m.conn.QueryRowCtx(ctx, &count, countQuery, args...)
	if err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	query := fmt.Sprintf("select %s from %s %s order by document_code limit %d, %d", aiConfigRows, m.table, whereCondition, offset, pageSize)
	var resp []*AiConfig
	err = m.conn.QueryRowsCtx(ctx, &resp, query, args...)
	return resp, count, err

}

func (m *customAiConfigModel) FindByNameAndCode(ctx context.Context, clientId string, name string, documentCode string) (*AiConfig, error) {
	var resp AiConfig
	var query string
	var err error

	if documentCode != "" {
		// 同时查询 name 和 documentCode
		query = fmt.Sprintf("select %s from %s where `client_id` = ? and `name` = ? and `document_code` = ? limit 1", aiConfigRows, m.table)
		err = m.conn.QueryRowCtx(ctx, &resp, query, clientId, name, documentCode)
	} else {
		// 只查询 name（如果表中有 document_code 字段，也可以查询 document_code IS NULL 的记录）
		query = fmt.Sprintf("select %s from %s where `client_id` = ? and `name` = ? and (`document_code` IS NULL OR `document_code` = '') limit 1", aiConfigRows, m.table)
		err = m.conn.QueryRowCtx(ctx, &resp, query, clientId, name)
	}

	switch err {
	case nil:
		return &resp, nil
	case sqlx.ErrNotFound:
		return nil, ErrNotFound
	default:
		return nil, err
	}
}

func (m *customAiConfigModel) FindListWithDocumentCode(ctx context.Context, clientId string, name string, documentCode string, page, pageSize uint64) ([]*AiConfig, int64, error) {
	whereConditions := []string{"client_id = ?"}
	args := []interface{}{clientId}

	if name != "" {
		keywordLike := "%" + name + "%"
		whereConditions = append(whereConditions, "name like ?")
		args = append(args, keywordLike)
	}

	if documentCode != "" {
		whereConditions = append(whereConditions, "document_code = ?")
		args = append(args, documentCode)
	}

	whereCondition := ""
	if len(whereConditions) > 0 {
		whereCondition = " where " + whereConditions[0]
		for i := 1; i < len(whereConditions); i++ {
			whereCondition = whereCondition + " and " + whereConditions[i]
		}
	}

	// 查询总数
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM %s %s", m.table, whereCondition)
	var count int64
	err := m.conn.QueryRowCtx(ctx, &count, countQuery, args...)
	if err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	query := fmt.Sprintf("select %s from %s %s order by id desc limit %d, %d", aiConfigRows, m.table, whereCondition, offset, pageSize)
	var resp []*AiConfig
	err = m.conn.QueryRowsCtx(ctx, &resp, query, args...)
	return resp, count, err
}

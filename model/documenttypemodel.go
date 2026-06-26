package model

import (
	"context"
	"fmt"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ DocumentTypeModel = (*customDocumentTypeModel)(nil)

type (
	// DocumentTypeModel is an interface to be customized, add more methods here,
	// and implement the added methods in customDocumentTypeModel.
	DocumentTypeModel interface {
		documentTypeModel
		withSession(session sqlx.Session) DocumentTypeModel
		FindAll(ctx context.Context) ([]*DocumentType, error)
		FindAllByClientId(ctx context.Context, clientId string) ([]*DocumentType, error)
		// FindOneByClientIdCode 按租户查询（禁止使用 gen 的 FindOne(code)）
		FindOneByClientIdCode(ctx context.Context, clientId, code string) (*DocumentType, error)
		// DeleteByClientIdCode 按租户删除（覆盖 gen Delete 仅按 code）
		DeleteByClientIdCode(ctx context.Context, clientId, code string) error
	}

	customDocumentTypeModel struct {
		*defaultDocumentTypeModel
	}
)

// NewDocumentTypeModel returns a model for the database table.
func NewDocumentTypeModel(conn sqlx.SqlConn) DocumentTypeModel {
	return &customDocumentTypeModel{
		defaultDocumentTypeModel: newDocumentTypeModel(conn),
	}
}

func (m *customDocumentTypeModel) withSession(session sqlx.Session) DocumentTypeModel {
	return NewDocumentTypeModel(sqlx.NewSqlConnFromSession(session))
}

func (m *customDocumentTypeModel) FindAll(ctx context.Context) ([]*DocumentType, error) {
	query := fmt.Sprintf("select %s from %s ORDER BY created_at DESC", documentTypeRows, m.table)

	// query := "SELECT code, name, description, created_at, updated_at, is_disabled FROM document_type ORDER BY created_at DESC"
	var documentTypes []*DocumentType
	err := m.conn.QueryRowsCtx(ctx, &documentTypes, query)
	return documentTypes, err
}

func (m *customDocumentTypeModel) FindAllByClientId(ctx context.Context, clientId string) ([]*DocumentType, error) {
	query := fmt.Sprintf("select %s from %s WHERE `client_id` = ? ORDER BY created_at DESC", documentTypeRows, m.table)
	var documentTypes []*DocumentType
	err := m.conn.QueryRowsCtx(ctx, &documentTypes, query, clientId)
	return documentTypes, err
}

func (m *customDocumentTypeModel) FindOneByClientIdCode(ctx context.Context, clientId, code string) (*DocumentType, error) {
	query := fmt.Sprintf("select %s from %s where `client_id` = ? and `code` = ? limit 1", documentTypeRows, m.table)
	var resp DocumentType
	err := m.conn.QueryRowCtx(ctx, &resp, query, clientId, code)
	switch err {
	case nil:
		return &resp, nil
	case sqlx.ErrNotFound:
		return nil, ErrNotFound
	default:
		return nil, err
	}
}

// Update 按租户更新：WHERE client_id + code（覆盖 gen 仅按 code）
func (m *customDocumentTypeModel) Update(ctx context.Context, data *DocumentType) error {
	query := fmt.Sprintf("update %s set %s where `client_id` = ? and `code` = ?", m.tableName(), documentTypeRowsWithPlaceHolder)
	_, err := m.conn.ExecCtx(ctx, query, data.ClientId, data.Name, data.IsDisabled, data.Description, data.ClientId, data.Code)
	return err
}

func (m *customDocumentTypeModel) DeleteByClientIdCode(ctx context.Context, clientId, code string) error {
	query := fmt.Sprintf("delete from %s where `client_id` = ? and `code` = ?", m.table)
	_, err := m.conn.ExecCtx(ctx, query, clientId, code)
	return err
}

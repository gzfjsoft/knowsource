package model

import (
	"context"
	"fmt"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ RawDocumentQaPairsModel = (*customRawDocumentQaPairsModel)(nil)

type (
	RawDocumentQaPairsModel interface {
		rawDocumentQaPairsModel
		DeleteByRawDocumentId(ctx context.Context, clientId string, rawDocumentId int64) error
		ListByRawDocumentId(ctx context.Context, clientId string, rawDocumentId int64, offset int64, limit int64) ([]*RawDocumentQaPairs, error)
		CountByRawDocumentId(ctx context.Context, clientId string, rawDocumentId int64) (int64, error)
	}

	customRawDocumentQaPairsModel struct {
		*defaultRawDocumentQaPairsModel
	}
)

func NewRawDocumentQaPairsModel(conn sqlx.SqlConn) RawDocumentQaPairsModel {
	return &customRawDocumentQaPairsModel{
		defaultRawDocumentQaPairsModel: newRawDocumentQaPairsModel(conn),
	}
}

func (m *customRawDocumentQaPairsModel) DeleteByRawDocumentId(ctx context.Context, clientId string, rawDocumentId int64) error {
	query := `DELETE FROM raw_document_qa_pairs WHERE client_id = ? AND raw_document_id = ?`
	_, err := m.conn.ExecCtx(ctx, query, clientId, rawDocumentId)
	return err
}

func (m *customRawDocumentQaPairsModel) ListByRawDocumentId(ctx context.Context, clientId string, rawDocumentId int64, offset int64, limit int64) ([]*RawDocumentQaPairs, error) {
	if limit <= 0 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}
	query := `SELECT id, client_id, raw_document_id, document_code, file_name, chunk_index, question, answer, qdrant_point_id, created_at, updated_at
		FROM raw_document_qa_pairs
		WHERE client_id = ? AND raw_document_id = ?
		ORDER BY chunk_index ASC, id ASC
		LIMIT ? OFFSET ?`
	out := make([]*RawDocumentQaPairs, 0)
	if err := m.conn.QueryRowsCtx(ctx, &out, query, clientId, rawDocumentId, limit, offset); err != nil {
		return nil, err
	}
	return out, nil
}

func (m *customRawDocumentQaPairsModel) CountByRawDocumentId(ctx context.Context, clientId string, rawDocumentId int64) (int64, error) {
	query := `SELECT COUNT(1) FROM raw_document_qa_pairs WHERE client_id = ? AND raw_document_id = ?`
	var total int64
	if err := m.conn.QueryRowCtx(ctx, &total, query, clientId, rawDocumentId); err != nil {
		return 0, fmt.Errorf("count raw_document_qa_pairs failed: %w", err)
	}
	return total, nil
}

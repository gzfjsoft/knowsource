package model

import (
	"context"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ DeptDocumentTypeModel = (*customDeptDocumentTypeModel)(nil)

type (
	// DeptDocumentTypeModel is an interface to be customized, add more methods here,
	// and implement the added methods in customDeptDocumentTypeModel.
	DeptDocumentTypeModel interface {
		deptDocumentTypeModel
		withSession(session sqlx.Session) DeptDocumentTypeModel
		// DeleteOrphansNotInDocumentType 删除 dept_document_type 中 (client_id, document_type_code)
		// 在 document_type 中无对应 (client_id, code) 的行。
		DeleteOrphansNotInDocumentType(ctx context.Context) (rowsAffected int64, err error)
	}

	customDeptDocumentTypeModel struct {
		*defaultDeptDocumentTypeModel
	}
)

// NewDeptDocumentTypeModel returns a model for the database table.
func NewDeptDocumentTypeModel(conn sqlx.SqlConn) DeptDocumentTypeModel {
	return &customDeptDocumentTypeModel{
		defaultDeptDocumentTypeModel: newDeptDocumentTypeModel(conn),
	}
}

func (m *customDeptDocumentTypeModel) withSession(session sqlx.Session) DeptDocumentTypeModel {
	return NewDeptDocumentTypeModel(sqlx.NewSqlConnFromSession(session))
}

func (m *customDeptDocumentTypeModel) DeleteOrphansNotInDocumentType(ctx context.Context) (int64, error) {
	const q = `DELETE ddt FROM dept_document_type ddt
LEFT JOIN document_type dt ON ddt.client_id = dt.client_id AND ddt.document_type_code = dt.code
WHERE dt.id IS NULL`
	res, err := m.conn.ExecCtx(ctx, q)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}

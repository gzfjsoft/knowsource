package model

import "github.com/zeromicro/go-zero/core/stores/sqlx"

var _ EmpDocumentTypeModel = (*customEmpDocumentTypeModel)(nil)

type (
	// EmpDocumentTypeModel is an interface to be customized, add more methods here,
	// and implement the added methods in customEmpDocumentTypeModel.
	EmpDocumentTypeModel interface {
		empDocumentTypeModel
		withSession(session sqlx.Session) EmpDocumentTypeModel
	}

	customEmpDocumentTypeModel struct {
		*defaultEmpDocumentTypeModel
	}
)

// NewEmpDocumentTypeModel returns a model for the database table.
func NewEmpDocumentTypeModel(conn sqlx.SqlConn) EmpDocumentTypeModel {
	return &customEmpDocumentTypeModel{
		defaultEmpDocumentTypeModel: newEmpDocumentTypeModel(conn),
	}
}

func (m *customEmpDocumentTypeModel) withSession(session sqlx.Session) EmpDocumentTypeModel {
	return NewEmpDocumentTypeModel(sqlx.NewSqlConnFromSession(session))
}

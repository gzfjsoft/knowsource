package model

import "github.com/zeromicro/go-zero/core/stores/sqlx"

var _ AdminAuditLogModel = (*customAdminAuditLogModel)(nil)

type (
	// AdminAuditLogModel is an interface to be customized, add more methods here,
	// and implement the added methods in customAdminAuditLogModel.
	AdminAuditLogModel interface {
		adminAuditLogModel
		withSession(session sqlx.Session) AdminAuditLogModel
	}

	customAdminAuditLogModel struct {
		*defaultAdminAuditLogModel
	}
)

// NewAdminAuditLogModel returns a model for the database table.
func NewAdminAuditLogModel(conn sqlx.SqlConn) AdminAuditLogModel {
	return &customAdminAuditLogModel{
		defaultAdminAuditLogModel: newAdminAuditLogModel(conn),
	}
}

func (m *customAdminAuditLogModel) withSession(session sqlx.Session) AdminAuditLogModel {
	return NewAdminAuditLogModel(sqlx.NewSqlConnFromSession(session))
}

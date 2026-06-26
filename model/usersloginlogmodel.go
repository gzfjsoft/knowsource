package model

import "github.com/zeromicro/go-zero/core/stores/sqlx"

var _ UsersLoginLogModel = (*customUsersLoginLogModel)(nil)

type (
	// UsersLoginLogModel is an interface to be customized, add more methods here,
	// and implement the added methods in customUsersLoginLogModel.
	UsersLoginLogModel interface {
		usersLoginLogModel
		withSession(session sqlx.Session) UsersLoginLogModel
	}

	customUsersLoginLogModel struct {
		*defaultUsersLoginLogModel
	}
)

// NewUsersLoginLogModel returns a model for the database table.
func NewUsersLoginLogModel(conn sqlx.SqlConn) UsersLoginLogModel {
	return &customUsersLoginLogModel{
		defaultUsersLoginLogModel: newUsersLoginLogModel(conn),
	}
}

func (m *customUsersLoginLogModel) withSession(session sqlx.Session) UsersLoginLogModel {
	return NewUsersLoginLogModel(sqlx.NewSqlConnFromSession(session))
}

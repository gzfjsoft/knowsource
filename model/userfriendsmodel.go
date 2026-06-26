package model

import "github.com/zeromicro/go-zero/core/stores/sqlx"

var _ UserFriendsModel = (*customUserFriendsModel)(nil)

type (
	// UserFriendsModel is an interface to be customized, add more methods here,
	// and implement the added methods in customUserFriendsModel.
	UserFriendsModel interface {
		userFriendsModel
		withSession(session sqlx.Session) UserFriendsModel
	}

	customUserFriendsModel struct {
		*defaultUserFriendsModel
	}
)

// NewUserFriendsModel returns a model for the database table.
func NewUserFriendsModel(conn sqlx.SqlConn) UserFriendsModel {
	return &customUserFriendsModel{
		defaultUserFriendsModel: newUserFriendsModel(conn),
	}
}

func (m *customUserFriendsModel) withSession(session sqlx.Session) UserFriendsModel {
	return NewUserFriendsModel(sqlx.NewSqlConnFromSession(session))
}

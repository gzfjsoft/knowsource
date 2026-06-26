package model

import (
	"context"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ FrPermissionsModel = (*customFrPermissionsModel)(nil)

type (
	// FrPermissionsModel is an interface to be customized, add more methods here,
	// and implement the added methods in customFrPermissionsModel.
	FrPermissionsModel interface {
		frPermissionsModel
		withSession(session sqlx.Session) FrPermissionsModel
		// DeleteByPermission 按 permission 删除（通过 FindOneByPermission + Delete(id)）
		DeleteByPermission(ctx context.Context, permission string) error
	}

	customFrPermissionsModel struct {
		*defaultFrPermissionsModel
	}
)

// NewFrPermissionsModel returns a model for the database table.
func NewFrPermissionsModel(conn sqlx.SqlConn) FrPermissionsModel {
	return &customFrPermissionsModel{
		defaultFrPermissionsModel: newFrPermissionsModel(conn),
	}
}

func (m *customFrPermissionsModel) withSession(session sqlx.Session) FrPermissionsModel {
	return NewFrPermissionsModel(sqlx.NewSqlConnFromSession(session))
}

func (m *customFrPermissionsModel) DeleteByPermission(ctx context.Context, permission string) error {
	row, err := m.FindOneByPermission(ctx, permission)
	if err != nil {
		return err
	}
	return m.Delete(ctx, row.Id)
}

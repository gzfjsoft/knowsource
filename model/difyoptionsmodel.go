package model

import (
	"context"
	"fmt"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ DifyOptionsModel = (*customDifyOptionsModel)(nil)

type (
	// DifyOptionsModel is an interface to be customized, add more methods here,
	// and implement the added methods in customDifyOptionsModel.
	DifyOptionsModel interface {
		difyOptionsModel
		withSession(session sqlx.Session) DifyOptionsModel
		// FindAllByClientId 按租户查询全部配置
		FindAllByClientId(ctx context.Context, clientId string) ([]*DifyOptions, error)
		// FindOneByClientIdName 按租户查询单条配置（禁止使用 gen 的 FindOne(name)）
		FindOneByClientIdName(ctx context.Context, clientId, name string) (*DifyOptions, error)
		// DeleteByClientIdName 按租户删除（覆盖 gen Delete 仅按 name）
		DeleteByClientIdName(ctx context.Context, clientId, name string) error
	}

	customDifyOptionsModel struct {
		*defaultDifyOptionsModel
	}
)

// NewDifyOptionsModel returns a model for the database table.
func NewDifyOptionsModel(conn sqlx.SqlConn) DifyOptionsModel {
	return &customDifyOptionsModel{
		defaultDifyOptionsModel: newDifyOptionsModel(conn),
	}
}

func (m *customDifyOptionsModel) withSession(session sqlx.Session) DifyOptionsModel {
	return NewDifyOptionsModel(sqlx.NewSqlConnFromSession(session))
}

func (m *customDifyOptionsModel) FindAllByClientId(ctx context.Context, clientId string) ([]*DifyOptions, error) {
	query := fmt.Sprintf("select %s from %s where `client_id` = ?", difyOptionsRows, m.table)
	var resp []*DifyOptions
	err := m.conn.QueryRowsCtx(ctx, &resp, query, clientId)
	return resp, err
}

func (m *customDifyOptionsModel) FindOneByClientIdName(ctx context.Context, clientId, name string) (*DifyOptions, error) {
	query := fmt.Sprintf("select %s from %s where `client_id` = ? and `name` = ? limit 1", difyOptionsRows, m.table)
	var resp DifyOptions
	err := m.conn.QueryRowCtx(ctx, &resp, query, clientId, name)
	switch err {
	case nil:
		return &resp, nil
	case sqlx.ErrNotFound:
		return nil, ErrNotFound
	default:
		return nil, err
	}
}

// Update 按租户更新：WHERE client_id + name（覆盖 gen 仅按 name）
func (m *customDifyOptionsModel) Update(ctx context.Context, data *DifyOptions) error {
	query := fmt.Sprintf("update %s set %s where `client_id` = ? and `name` = ?", m.table, difyOptionsRowsWithPlaceHolder)
	_, err := m.conn.ExecCtx(ctx, query, data.ClientId, data.Url, data.ApiKey, data.Description, data.ClientId, data.Name)
	return err
}

func (m *customDifyOptionsModel) DeleteByClientIdName(ctx context.Context, clientId, name string) error {
	query := fmt.Sprintf("delete from %s where `client_id` = ? and `name` = ?", m.table)
	_, err := m.conn.ExecCtx(ctx, query, clientId, name)
	return err
}

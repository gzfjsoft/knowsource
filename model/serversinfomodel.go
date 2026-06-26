package model

import (
	"context"
	"fmt"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ ServersInfoModel = (*customServersInfoModel)(nil)

type (
	// ServersInfoModel is an interface to be customized, add more methods here,
	// and implement the added methods in customServersInfoModel.
	ServersInfoModel interface {
		serversInfoModel
		withSession(session sqlx.Session) ServersInfoModel
		FindAll(ctx context.Context, page int64, pageSize int64) ([]*ServersInfo, error)
		FindOneWithName(ctx context.Context, name string) (*ServersInfo, error)
		Count(ctx context.Context) (int64, error)
	}

	customServersInfoModel struct {
		*defaultServersInfoModel
	}
)

// NewServersInfoModel returns a model for the database table.
func NewServersInfoModel(conn sqlx.SqlConn) ServersInfoModel {
	return &customServersInfoModel{
		defaultServersInfoModel: newServersInfoModel(conn),
	}
}

func (m *customServersInfoModel) withSession(session sqlx.Session) ServersInfoModel {
	return NewServersInfoModel(sqlx.NewSqlConnFromSession(session))
}

func (m *customServersInfoModel) FindOneWithName(ctx context.Context, name string) (*ServersInfo, error) {
	query := fmt.Sprintf("select * from %s where server_name = ?", m.table)
	var resp ServersInfo
	err := m.conn.QueryRowCtx(ctx, &resp, query, name)
	return &resp, err
}

func (m *customServersInfoModel) FindAll(ctx context.Context, page int64, pageSize int64) ([]*ServersInfo, error) {

	offset := (page - 1) * pageSize

	query := fmt.Sprintf("select * from %s limit %d, %d", m.table, offset, pageSize)
	var resp []*ServersInfo
	err := m.conn.QueryRowsCtx(ctx, &resp, query)
	return resp, err
}

func (m *customServersInfoModel) Count(ctx context.Context) (int64, error) {
	query := fmt.Sprintf("select count(*) from %s", m.table)
	var count int64
	err := m.conn.QueryRowCtx(ctx, &count, query)
	return count, err
}

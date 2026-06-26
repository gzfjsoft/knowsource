package model

import (
	"context"
	"fmt"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ ServersModel = (*customServersModel)(nil)

type (
	// ServersModel is an interface to be customized, add more methods here,
	// and implement the added methods in customServersModel.
	ServersModel interface {
		serversModel
		WithSession(session sqlx.Session) ServersModel
		FindList(ctx context.Context, condition string) (*[]Servers, error)
		Count(ctx context.Context, condition string) (int, error)
		FindOneWithName(ctx context.Context, serverName string) (*Servers, error)
		//DecreaseGpuUsed(ctx context.Context, serverId uint64, gpuUsed int64) error
		IncreaseGpuUsed(ctx context.Context, serverId uint64, gpuUsed int64) error
		RecalculateServerGpuCount(ctx context.Context, serverId uint64) error
	}

	customServersModel struct {
		*defaultServersModel
	}
)

// NewServersModel returns a model for the database table.
func NewServersModel(conn sqlx.SqlConn) ServersModel {
	return &customServersModel{
		defaultServersModel: newServersModel(conn),
	}
}

func (m *customServersModel) WithSession(session sqlx.Session) ServersModel {
	return NewServersModel(sqlx.NewSqlConnFromSession(session))
}

// func (m *customServersModel) DecreaseGpuUsed(ctx context.Context, serverId uint64, gpuUsed int64) error {
// 	query := fmt.Sprintf("update %s set gpu_used = gpu_used - ? where server_id = ?", m.table)
// 	_, err := m.conn.ExecCtx(ctx, query, gpuUsed, serverId)
// 	return err
// }

func (m *customServersModel) RecalculateServerGpuCount(ctx context.Context, serverId uint64) error {
	query := `UPDATE servers JOIN ( SELECT server_id, SUM(gpu_cores) as sum_gpu    FROM instances     WHERE state = 'started'     GROUP BY server_id ) AS data 
	 ON servers.server_id = data.server_id SET servers.gpu_used = data.sum_gpu where servers.server_id = ?`
	query2 := `UPDATE servers SET gpu_used = 0 WHERE server_id not in (select distinct server_id from instances where state="started") and server_id = ?`
	query3 := `update servers set gpu_used = gpu_count  where gpu_count < gpu_used and server_id = ?`

	_, err := m.conn.ExecCtx(ctx, query, serverId)
	if err != nil {
		return err
	}
	_, err = m.conn.ExecCtx(ctx, query3, serverId)
	if err != nil {
		return err
	}
	_, err = m.conn.ExecCtx(ctx, query2, serverId)
	if err != nil {
		return err
	}
	return nil
}

func (m *customServersModel) IncreaseGpuUsed(ctx context.Context, serverId uint64, gpuUsed int64) error {
	query := fmt.Sprintf("update %s set gpu_used = gpu_used + ? where server_id = ?", m.table)
	_, err := m.conn.ExecCtx(ctx, query, gpuUsed, serverId)
	return err
}

func (m *customServersModel) FindOneWithName(ctx context.Context, serverName string) (*Servers, error) {
	query := fmt.Sprintf("select %s from %s where server_name = ?", serversRows, m.table)
	var resp Servers
	err := m.conn.QueryRowCtx(ctx, &resp, query, serverName)
	switch err {
	case nil:
		return &resp, nil
	default:
		return nil, err
	}
}

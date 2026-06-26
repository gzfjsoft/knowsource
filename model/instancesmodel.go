package model

import (
	"context"
	"fmt"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ InstancesModel = (*customInstancesModel)(nil)

type (
	// InstancesModel is an interface to be customized, add more methods here,
	// and implement the added methods in customInstancesModel.
	InstancesModel interface {
		instancesModel
		WithSession(session sqlx.Session) InstancesModel
		FindByFilter(ctx context.Context, orgId uint64, userId, serverId uint64, state string, page, pageSize uint64) ([]*Instances, uint64, error)
		FindAllRunning(ctx context.Context) ([]*Instances, error)
		FindAll(ctx context.Context, condition string) ([]*Instances, error)
		Count(ctx context.Context, condition string) (uint64, error)
		FindByName(ctx context.Context, name string) (*Instances, error)
	}

	customInstancesModel struct {
		*defaultInstancesModel
	}
)

// NewInstancesModel returns a model for the database table.
func NewInstancesModel(conn sqlx.SqlConn) InstancesModel {
	return &customInstancesModel{
		defaultInstancesModel: newInstancesModel(conn),
	}
}

func (m *customInstancesModel) WithSession(session sqlx.Session) InstancesModel {
	return NewInstancesModel(sqlx.NewSqlConnFromSession(session))
}

func (m *customInstancesModel) FindByName(ctx context.Context, name string) (*Instances, error) {
	query := fmt.Sprintf("select %s from %s where name = ? limit 1", instancesRows, m.table)
	var resp *Instances
	err := m.conn.QueryRowCtx(ctx, &resp, query, name)
	return resp, err
}

func (m *customInstancesModel) FindAll(ctx context.Context, condition string) ([]*Instances, error) {
	query := fmt.Sprintf("select %s from %s where %s", instancesRows, m.table, condition)
	var resp []*Instances
	err := m.conn.QueryRowsCtx(ctx, &resp, query)
	return resp, err
}

func (m *customInstancesModel) FindAllRunning(ctx context.Context) ([]*Instances, error) {
	query := fmt.Sprintf("select %s from %s where state = 'running'", instancesRows, m.table)
	var resp []*Instances
	err := m.conn.QueryRowsCtx(ctx, &resp, query)
	return resp, err
}

func (m *customInstancesModel) FindByFilter(ctx context.Context, orgId uint64, userId, serverId uint64, state string, page, pageSize uint64) ([]*Instances, uint64, error) {
	where := "1=1"
	var args []interface{}

	if orgId > 0 {
		where += " AND org_id = ?"
		args = append(args, orgId)
	}

	if userId > 0 {
		where += " AND user_id = ?"
		args = append(args, userId)
	}

	if serverId > 0 {
		where += " AND server_id = ?"
		args = append(args, serverId)
	}
	if state != "" {
		where += " AND state = ?"
		args = append(args, state)
	}

	query := fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE %s", m.table, where)
	var total uint64
	err := m.conn.QueryRowCtx(ctx, &total, query, args...)
	if err != nil {
		return nil, 0, err
	}

	if total == 0 {
		return []*Instances{}, 0, nil
	}

	query = fmt.Sprintf("SELECT %s FROM %s WHERE %s ORDER BY instance_id DESC LIMIT ?,?", instancesRows, m.table, where)
	args = append(args, (page-1)*pageSize, pageSize)
	var resp []*Instances
	err = m.conn.QueryRowsCtx(ctx, &resp, query, args...)
	if err != nil {
		return nil, 0, err
	}

	return resp, total, nil
}

func (m *customInstancesModel) Count(ctx context.Context, condition string) (uint64, error) {
	query := fmt.Sprintf("select count(*) from %s where %s", m.table, condition)
	var count uint64
	err := m.conn.QueryRowCtx(ctx, &count, query)
	return count, err
}

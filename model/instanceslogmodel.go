package model

import (
	"context"
	"fmt"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ InstancesLogModel = (*customInstancesLogModel)(nil)

type (
	// InstancesLogModel is an interface to be customized, add more methods here,
	// and implement the added methods in customInstancesLogModel.
	InstancesLogModel interface {
		instancesLogModel
		withSession(session sqlx.Session) InstancesLogModel
		FindAll(ctx context.Context, instanceID uint64) ([]*InstancesLog, error)
	}

	customInstancesLogModel struct {
		*defaultInstancesLogModel
	}
)

// NewInstancesLogModel returns a model for the database table.
func NewInstancesLogModel(conn sqlx.SqlConn) InstancesLogModel {
	return &customInstancesLogModel{
		defaultInstancesLogModel: newInstancesLogModel(conn),
	}
}

func (m *customInstancesLogModel) withSession(session sqlx.Session) InstancesLogModel {
	return NewInstancesLogModel(sqlx.NewSqlConnFromSession(session))
}

func (m *customInstancesLogModel) FindAll(ctx context.Context, instanceID uint64) ([]*InstancesLog, error) {
	query := fmt.Sprintf("SELECT * FROM %s WHERE instance_id = ? ", m.table)
	var resp []*InstancesLog
	err := m.conn.QueryRowsCtx(ctx, &resp, query, instanceID)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

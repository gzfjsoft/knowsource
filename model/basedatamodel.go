package model

import (
	"context"
	"fmt"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ BaseDataModel = (*customBaseDataModel)(nil)

type (
	// BaseDataModel is an interface to be customized, add more methods here,
	// and implement the added methods in customBaseDataModel.
	BaseDataModel interface {
		baseDataModel
		withSession(session sqlx.Session) BaseDataModel

		FindByType(ctx context.Context, datatype string) (*[]BaseData, error)
		FindList(ctx context.Context, condition string) (*[]BaseData, error)
		Count(ctx context.Context, condition string) (int, error)
		FindByTypeName(ctx context.Context, datatype string, name string) (int, error)
	}

	customBaseDataModel struct {
		*defaultBaseDataModel
	}
)

// NewBaseDataModel returns a model for the database table.
func NewBaseDataModel(conn sqlx.SqlConn) BaseDataModel {
	return &customBaseDataModel{
		defaultBaseDataModel: newBaseDataModel(conn),
	}
}

func (m *customBaseDataModel) withSession(session sqlx.Session) BaseDataModel {
	return NewBaseDataModel(sqlx.NewSqlConnFromSession(session))
}

func (m *customBaseDataModel) FindByType(ctx context.Context, datatype string) (*[]BaseData, error) {
	query := fmt.Sprintf("select %s from %s where `data_type` = ? order by value", baseDataRows, m.table)
	var resp []BaseData
	err := m.conn.QueryRowsCtx(ctx, &resp, query, datatype)
	switch err {
	case nil:
		return &resp, nil
	case sqlx.ErrNotFound:
		return nil, ErrNotFound
	default:
		return nil, err
	}
}

func (m *customBaseDataModel) FindByTypeName(ctx context.Context, datatype string, value string) (int, error) {
	query := fmt.Sprintf("select `id` from %s where `data_type` = ? and `name` = ?", m.table)
	count := 0
	err := m.conn.QueryRowCtx(ctx, &count, query, datatype, value)
	if err != nil {
		return 0, err
	}
	return count, nil
}

func (m *customBaseDataModel) FindList(ctx context.Context, condition string) (*[]BaseData, error) {
	query := fmt.Sprintf("select %s from %s %s", baseDataRows, m.table, condition)
	var resp []BaseData
	err := m.conn.QueryRowsCtx(ctx, &resp, query)
	switch err {
	case nil:
		return &resp, nil
	case sqlx.ErrNotFound:
		return nil, ErrNotFound
	default:
		return nil, err
	}
}

func (m *customBaseDataModel) Count(ctx context.Context, condition string) (int, error) {
	query := fmt.Sprintf("select count(1) from %s %s", m.table, condition)
	count := 0
	err := m.conn.QueryRowCtx(ctx, &count, query)
	if err != nil {
		return 0, err
	}
	return count, nil
}

package model

import (
	"context"
	"fmt"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ MaintainInfoModel = (*customMaintainInfoModel)(nil)

type (
	// MaintainInfoModel is an interface to be customized, add more methods here,
	// and implement the added methods in customMaintainInfoModel.
	MaintainInfoModel interface {
		maintainInfoModel
		withSession(session sqlx.Session) MaintainInfoModel
		FindOneByMd5(ctx context.Context, md5 string) (*MaintainInfo, error)
		FindAll(ctx context.Context, typ string, offset, limit int64) ([]*MaintainInfo, error)
		Search(ctx context.Context, keyword, typ string, offset, limit int64) ([]*MaintainInfo, error)
		Count(ctx context.Context, typ string) (int64, error)
		CountSearch(ctx context.Context, keyword, typ string) (int64, error)
	}

	customMaintainInfoModel struct {
		*defaultMaintainInfoModel
	}
)

// NewMaintainInfoModel returns a model for the database table.
func NewMaintainInfoModel(conn sqlx.SqlConn) MaintainInfoModel {
	return &customMaintainInfoModel{
		defaultMaintainInfoModel: newMaintainInfoModel(conn),
	}
}

func (m *customMaintainInfoModel) withSession(session sqlx.Session) MaintainInfoModel {
	return NewMaintainInfoModel(sqlx.NewSqlConnFromSession(session))
}

func (m *customMaintainInfoModel) FindOneByMd5(ctx context.Context, md5 string) (*MaintainInfo, error) {
	query := fmt.Sprintf("select %s from %s where `md5` = ? limit 1", maintainInfoRows, m.table)
	var resp MaintainInfo
	err := m.conn.QueryRowCtx(ctx, &resp, query, md5)
	switch err {
	case nil:
		return &resp, nil
	case sqlx.ErrNotFound:
		return nil, ErrNotFound
	default:
		return nil, err
	}
}

func (m *customMaintainInfoModel) FindAll(ctx context.Context, typ string, offset, limit int64) ([]*MaintainInfo, error) {
	query := fmt.Sprintf("select %s from %s", maintainInfoRows, m.table)
	args := []interface{}{}
	if typ != "" {
		query += " where type = ?"
		args = append(args, typ)
	}
	query += " order by id desc limit ? offset ?"
	args = append(args, limit, offset)
	var resp []*MaintainInfo
	err := m.conn.QueryRowsCtx(ctx, &resp, query, args...)
	return resp, err
}

func (m *customMaintainInfoModel) Search(ctx context.Context, keyword, typ string, offset, limit int64) ([]*MaintainInfo, error) {
	query := fmt.Sprintf("select %s from %s where 1=1", maintainInfoRows, m.table)
	args := []interface{}{}
	if typ != "" {
		query += " and type = ?"
		args = append(args, typ)
	}
	if keyword != "" {
		query += " and (title like ? or info like ?)"
		args = append(args, "%"+keyword+"%", "%"+keyword+"%")
	}
	query += " order by id desc limit ? offset ?"
	args = append(args, limit, offset)
	var resp []*MaintainInfo
	err := m.conn.QueryRowsCtx(ctx, &resp, query, args...)
	return resp, err
}

func (m *customMaintainInfoModel) Count(ctx context.Context, typ string) (int64, error) {
	query := fmt.Sprintf("select count(*) from %s", m.table)
	args := []interface{}{}
	if typ != "" {
		query += " where type = ?"
		args = append(args, typ)
	}
	var count int64
	err := m.conn.QueryRowCtx(ctx, &count, query, args...)
	return count, err
}

func (m *customMaintainInfoModel) CountSearch(ctx context.Context, keyword, typ string) (int64, error) {
	query := fmt.Sprintf("select count(*) from %s where 1=1", m.table)
	args := []interface{}{}
	if typ != "" {
		query += " and type = ?"
		args = append(args, typ)
	}
	if keyword != "" {
		query += " and (title like ? or info like ?)"
		args = append(args, "%"+keyword+"%", "%"+keyword+"%")
	}
	var count int64
	err := m.conn.QueryRowCtx(ctx, &count, query, args...)
	return count, err
}

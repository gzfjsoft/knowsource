package model

import (
	"context"
	"fmt"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ BucketsModel = (*customBucketsModel)(nil)

type (
	// BucketsModel is an interface to be customized, add more methods here,
	// and implement the added methods in customBucketsModel.
	BucketsModel interface {
		bucketsModel
		withSession(session sqlx.Session) BucketsModel
		Count(ctx context.Context, region string, uid uint64) (uint64, error)
		FindMany(ctx context.Context, region string, uid uint64, offset, limit uint64) ([]*Buckets, error)
		FindByName(ctx context.Context, bucketName string) (*Buckets, error)
		FindById(ctx context.Context, bucketId uint64) (*Buckets, error)
	}

	customBucketsModel struct {
		*defaultBucketsModel
	}
)

// NewBucketsModel returns a model for the database table.
func NewBucketsModel(conn sqlx.SqlConn) BucketsModel {
	return &customBucketsModel{
		defaultBucketsModel: newBucketsModel(conn),
	}
}

func (m *customBucketsModel) withSession(session sqlx.Session) BucketsModel {
	return NewBucketsModel(sqlx.NewSqlConnFromSession(session))
}

func (m *customBucketsModel) FindByName(ctx context.Context, bucketName string) (*Buckets, error) {
	query := fmt.Sprintf("select %s from %s where `bucket_name` = ? and `is_deleted` = 0 limit 1", bucketsRows, m.table)
	var resp Buckets
	err := m.conn.QueryRowCtx(ctx, &resp, query, bucketName)
	switch err {
	case nil:
		return &resp, nil
	case sqlx.ErrNotFound:
		return nil, ErrNotFound
	default:
		return nil, err
	}
}

func (m *customBucketsModel) FindById(ctx context.Context, bucketId uint64) (*Buckets, error) {
	query := fmt.Sprintf("select %s from %s where `bucket_id` = ? and `is_deleted` = 0 limit 1", bucketsRows, m.table)
	var resp Buckets
	err := m.conn.QueryRowCtx(ctx, &resp, query, bucketId)
	switch err {
	case nil:
		return &resp, nil
	case sqlx.ErrNotFound:
		return nil, ErrNotFound
	default:
		return nil, err
	}
}

func (m *customBucketsModel) Count(ctx context.Context, region string, uid uint64) (uint64, error) {
	var count uint64
	query := fmt.Sprintf("select count(*) from %s where `is_deleted` = 0 and `user_id` = ?", m.table)

	if region != "" {
		query += " and `region` = ?"
		err := m.conn.QueryRowCtx(ctx, &count, query, uid, region)
		return count, err
	}
	err := m.conn.QueryRowCtx(ctx, &count, query, uid)
	return count, err
}

func (m *customBucketsModel) FindMany(ctx context.Context, region string, uid uint64, offset, limit uint64) ([]*Buckets, error) {
	query := fmt.Sprintf("select %s from %s where `is_deleted` = 0 and `user_id` = ?", bucketsRows, m.table)

	if region != "" {
		query += " and `region` = ?"
	}
	query += " order by `bucket_id` desc limit ?,?"

	var resp []*Buckets
	if region != "" {
		err := m.conn.QueryRowsCtx(ctx, &resp, query, uid, region, offset, limit)
		return resp, err
	}
	err := m.conn.QueryRowsCtx(ctx, &resp, query, uid, offset, limit)
	return resp, err
}

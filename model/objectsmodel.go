package model

import (
	"context"
	"fmt"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ ObjectsModel = (*customObjectsModel)(nil)

type (
	// ObjectsModel is an interface to be customized, add more methods here,
	// and implement the added methods in customObjectsModel.
	ObjectsModel interface {
		objectsModel
		WithSession(session sqlx.Session) ObjectsModel
		Count(ctx context.Context, bucketId uint64, treeId uint64, key string, userId uint64) (uint64, error)
		FindMany(ctx context.Context, bucketId uint64, TreeId uint64, key string, userId uint64, offset, limit uint64) ([]*Objects, error)
		FindByBucketAndKey(ctx context.Context, bucketId uint64, key string) (*Objects, error)
		FindByBucketId(ctx context.Context, bucketId uint64) ([]*Objects, error)
		DeleteByUserId(ctx context.Context, objId uint64, userId uint64) error
	}

	customObjectsModel struct {
		*defaultObjectsModel
	}
)

// NewObjectsModel returns a model for the database table.
func NewObjectsModel(conn sqlx.SqlConn) ObjectsModel {
	return &customObjectsModel{
		defaultObjectsModel: newObjectsModel(conn),
	}
}

func (m *customObjectsModel) WithSession(session sqlx.Session) ObjectsModel {
	return NewObjectsModel(sqlx.NewSqlConnFromSession(session))
}

func (m *customObjectsModel) DeleteByUserId(ctx context.Context, objId uint64, userId uint64) error {
	query := fmt.Sprintf("delete from %s where `obj_id` = ? and `user_id` = ?", m.table)
	_, err := m.conn.ExecCtx(ctx, query, objId, userId)
	return err
}

func (m *customObjectsModel) FindByBucketAndKey(ctx context.Context, bucketId uint64, key string) (*Objects, error) {
	query := fmt.Sprintf("select %s from %s where `bucket_id` = ? and `Key` = ? limit 1", objectsRows, m.table)
	var resp Objects
	err := m.conn.QueryRowCtx(ctx, &resp, query, bucketId, key)
	switch err {
	case nil:
		return &resp, nil
	case sqlx.ErrNotFound:
		return nil, ErrNotFound
	default:
		return nil, err
	}
}

func (m *customObjectsModel) Count(ctx context.Context, bucketId uint64, treeId uint64, key string, userId uint64) (uint64, error) {
	var count uint64
	query := fmt.Sprintf("select count(*) from %s where 1=1", m.table)

	var conditions []interface{}
	if bucketId > 0 {
		query += " and `bucket_id` = ?"
		conditions = append(conditions, bucketId)
	}
	if treeId > 0 {
		query += " and `tree_id` = ?"
		conditions = append(conditions, treeId)
	}
	if key != "" {
		query += " and `Key` like ?"
		conditions = append(conditions, "%"+key+"%")
	}
	if userId > 0 {
		query += " and `user_id` = ?"
		conditions = append(conditions, userId)
	}
	err := m.conn.QueryRowCtx(ctx, &count, query, conditions...)
	return count, err
}

func (m *customObjectsModel) FindMany(ctx context.Context, bucketId uint64, TreeId uint64, key string, userId uint64, offset, limit uint64) ([]*Objects, error) {
	query := fmt.Sprintf("select %s from %s where 1=1", objectsRows, m.table)

	var conditions []interface{}
	if bucketId > 0 {
		query += " and `bucket_id` = ?"
		conditions = append(conditions, bucketId)
	}
	if TreeId > 0 {
		query += " and `tree_id` = ?"
		conditions = append(conditions, TreeId)
	}
	if key != "" {
		query += " and `Key` like ?"
		conditions = append(conditions, "%"+key+"%")
	}
	if userId > 0 {
		query += " and `user_id` = ?"
		conditions = append(conditions, userId)
	}

	query += " order by `obj_id` desc limit ?,?"
	conditions = append(conditions, offset, limit)

	var resp []*Objects
	err := m.conn.QueryRowsCtx(ctx, &resp, query, conditions...)
	return resp, err
}

// FindByBucketId returns all objects in a bucket
func (m *customObjectsModel) FindByBucketId(ctx context.Context, bucketId uint64) ([]*Objects, error) {
	var resp []*Objects
	query := fmt.Sprintf("select %s from %s where `bucket_id` = ?", objectsRows, m.table)
	err := m.conn.QueryRowsCtx(ctx, &resp, query, bucketId)
	switch err {
	case nil:
		return resp, nil
	case sqlx.ErrNotFound:
		return nil, ErrNotFound
	default:
		return nil, err
	}
}

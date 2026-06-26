package model

import (
	"context"
	"fmt"

	// "github.com/go-redis/redis/v8"

	"github.com/zeromicro/go-zero/core/stores/redis"

	"github.com/vmihailenco/msgpack/v5"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ UserBlackListModel = (*customUserBlackListModel)(nil)

type (
	// UserBlackListModel is an interface to be customized, add more methods here,
	// and implement the added methods in customUserBlackListModel.
	UserBlackListModel interface {
		userBlackListModel
		withSession(session sqlx.Session) UserBlackListModel
		FindByUserId(ctx context.Context, userId uint64) ([]*UserBlackList, error)
		FindByBlackId(ctx context.Context, userId uint64) ([]*UserBlackList, error)
		FindByUserIdWithCache(ctx context.Context, rdb *redis.Redis, userId uint64) ([]*UserBlackList, error)
		FindByBlackIdWithCache(ctx context.Context, rdb *redis.Redis, userId uint64) ([]*UserBlackList, error)
		InvalidateBlacklistCache(rdb *redis.Redis, userId uint64) error
		CountByUserId(ctx context.Context, userId uint64) (uint64, error)
	}

	customUserBlackListModel struct {
		*defaultUserBlackListModel
	}
)

// NewUserBlackListModel returns a model for the database table.
func NewUserBlackListModel(conn sqlx.SqlConn) UserBlackListModel {
	return &customUserBlackListModel{
		defaultUserBlackListModel: newUserBlackListModel(conn),
	}
}

func (m *customUserBlackListModel) withSession(session sqlx.Session) UserBlackListModel {
	return NewUserBlackListModel(sqlx.NewSqlConnFromSession(session))
}

// InvalidateBlacklistCache invalidates the blacklist cache for a user
func (m *customUserBlackListModel) InvalidateBlacklistCache(rdb *redis.Redis, userId uint64) error {
	key := fmt.Sprintf("user_blacklist:user_id:%d", userId)

	_, err := rdb.Del(key)

	return err
}

// 保存用户黑名单到redis
func saveUserAsMsgPack(rdb *redis.Redis, key string, list []*UserBlackList) error {
	data, err := msgpack.Marshal(list)
	if err != nil {
		return err
	}
	// logx.Infof("saveUserAsMsgPack: key: %s, data: %s", key, string(data))
	return rdb.Setex(key, string(data), 30) //redis_timeout
}

// 从redis中获取用户黑名单
func getUserFromMsgPack(rdb *redis.Redis, key string) ([]*UserBlackList, error) {
	val, err := rdb.Get(key)
	if err != nil {
		return nil, err
	}
	var list []*UserBlackList
	err = msgpack.Unmarshal([]byte(val), &list)
	return list, err
}

func (m *customUserBlackListModel) FindByUserIdWithCache(ctx context.Context, rdb *redis.Redis, userId uint64) ([]*UserBlackList, error) {
	key := fmt.Sprintf("user_blacklist:user_id:%d", userId)
	list, err := getUserFromMsgPack(rdb, key)
	if err != nil {
		list, err = m.FindByUserId(ctx, userId)
		if err != nil {
			return nil, err
		}
		err = saveUserAsMsgPack(rdb, key, list)
		if err != nil {
			return nil, err
		}
	}
	return list, nil
}
func (m *customUserBlackListModel) FindByBlackIdWithCache(ctx context.Context, rdb *redis.Redis, userId uint64) ([]*UserBlackList, error) {
	key := fmt.Sprintf("user_blacklist:black_id:%d", userId)
	list, err := getUserFromMsgPack(rdb, key)
	if err != nil {
		list, err = m.FindByBlackId(ctx, userId)
		if err != nil {
			return nil, err
		}
		err = saveUserAsMsgPack(rdb, key, list)
		if err != nil {
			return nil, err
		}
	}
	return list, err
}

func (m *customUserBlackListModel) FindByUserId(ctx context.Context, userId uint64) ([]*UserBlackList, error) {
	where := fmt.Sprintf("where user_id = %d", userId)

	query := fmt.Sprintf("select %s from %s %s order by id desc", userBlackListRows, m.table, where)
	var resp []*UserBlackList
	err := m.conn.QueryRowsCtx(ctx, &resp, query)
	return resp, err
}
func (m *customUserBlackListModel) FindByBlackId(ctx context.Context, userId uint64) ([]*UserBlackList, error) {
	where := fmt.Sprintf("where black_id = %d", userId)

	query := fmt.Sprintf("select %s from %s %s order by id desc", userBlackListRows, m.table, where)
	var resp []*UserBlackList
	err := m.conn.QueryRowsCtx(ctx, &resp, query)
	return resp, err
}
func (m *customUserBlackListModel) CountByUserId(ctx context.Context, userId uint64) (uint64, error) {
	query := fmt.Sprintf("select count(*) from %s where user_id = %d", m.table, userId)
	var count uint64
	err := m.conn.QueryRowCtx(ctx, &count, query)
	return count, err
}

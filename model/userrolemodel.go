package model

import (
	"context"
	"fmt"

	"github.com/vmihailenco/msgpack/v5"
	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ UserRoleModel = (*customUserRoleModel)(nil)

type (
	// UserRoleModel is an interface to be customized, add more methods here,
	// and implement the added methods in customUserRoleModel.
	UserRoleModel interface {
		userRoleModel
		CountByRoleId(ctx context.Context, roleId int64) (int64, error)
		DeleteByUserId(ctx context.Context, userId int64) error
		DeleteByRoleId(ctx context.Context, roleId int64) error
		FindByUserIdWithCache(ctx context.Context, userId int64, rdb *redis.Redis) ([]*UserRole, error)
		FindAllWithCache(ctx context.Context, rdb *redis.Redis) ([]*UserRole, error)
		RemoveCache(ctx context.Context, rdb *redis.Redis) error
		WithSession(session sqlx.Session) UserRoleModel
	}

	customUserRoleModel struct {
		*defaultUserRoleModel
	}
)

// NewUserRoleModel returns a model for the database table.
func NewUserRoleModel(conn sqlx.SqlConn) UserRoleModel {
	return &customUserRoleModel{
		defaultUserRoleModel: newUserRoleModel(conn),
	}
}

func (m *customUserRoleModel) WithSession(session sqlx.Session) UserRoleModel {
	return NewUserRoleModel(sqlx.NewSqlConnFromSession(session))
}

func (m *customUserRoleModel) CountByRoleId(ctx context.Context, roleId int64) (int64, error) {
	var count int64
	err := m.conn.QueryRowCtx(ctx, &count, "SELECT COUNT(*) FROM user_role WHERE role_id = ?", roleId)
	if err != nil {
		return 0, err
	}
	return count, nil
}

func (m *customUserRoleModel) DeleteByUserId(ctx context.Context, userId int64) error {
	query := fmt.Sprintf("delete from %s where `user_id` = ?", m.table)
	_, err := m.conn.ExecCtx(ctx, query, userId)
	return err
}

func (m *customUserRoleModel) DeleteByRoleId(ctx context.Context, roleId int64) error {
	query := fmt.Sprintf("delete from %s where `role_id` = ?", m.table)
	_, err := m.conn.ExecCtx(ctx, query, roleId)
	return err
}

func (m *customUserRoleModel) FindByUserIdWithCache(ctx context.Context, userId int64, rdb *redis.Redis) ([]*UserRole, error) {
	allUserRoles, err := m.FindAllWithCache(ctx, rdb)
	if err != nil {
		return nil, err
	}
	var userRoles []*UserRole
	for _, userRole := range allUserRoles {
		if userRole.UserId == userId {
			userRoles = append(userRoles, userRole)
		}
	}
	return userRoles, nil
}

func (m *customUserRoleModel) FindAllWithCache(ctx context.Context, rdb *redis.Redis) ([]*UserRole, error) {
	key := "knowdata_user_role:all"
	userRoles, err := m.getUserRolesFromMsgPack(rdb, key)
	if err != nil {
		query := fmt.Sprintf("select %s from %s", userRoleRows, m.table)
		err := m.conn.QueryRowsCtx(ctx, &userRoles, query)
		if err != nil {
			return nil, err
		}
		// Update cache in background
		go func() {
			_ = m.saveUserRolesAsMsgPack(rdb, key, userRoles)
		}()
	}
	return userRoles, nil
}

func (m *customUserRoleModel) RemoveCache(ctx context.Context, rdb *redis.Redis) error {
	key := "knowdata_user_role:all"
	_, err := rdb.Del(key)
	return err
}

// saveUserRolesAsMsgPack saves user roles to redis using msgpack encoding
func (m *customUserRoleModel) saveUserRolesAsMsgPack(rdb *redis.Redis, key string, userRoles []*UserRole) error {
	data, err := msgpack.Marshal(userRoles)
	if err != nil {
		return err
	}
	err = rdb.Setex(key, string(data), 3600)
	return err
}

// getUserRolesFromMsgPack retrieves user roles from redis using msgpack encoding
func (m *customUserRoleModel) getUserRolesFromMsgPack(rdb *redis.Redis, key string) ([]*UserRole, error) {
	data, err := rdb.Get(key)
	if err != nil {
		return nil, err
	}
	var userRoles []*UserRole
	err = msgpack.Unmarshal([]byte(data), &userRoles)
	return userRoles, err
}

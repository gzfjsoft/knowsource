package model

import (
	"context"
	"fmt"
	"strings"

	"github.com/vmihailenco/msgpack/v5"
	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type RolePermissionData struct {
	RoleId int64  `db:"role_id"` // '角色id'
	Code   string `db:"code"`    // '权限标识'
}

var _ PermissionModel = (*customPermissionModel)(nil)

type (
	// PermissionModel is an interface to be customized, add more methods here,
	// and implement the added methods in customPermissionModel.
	PermissionModel interface {
		permissionModel
		CheckPermissionCodeExists(ctx context.Context, code string, id int64) (bool, error)
		FindByRoleIds(ctx context.Context, roleIds []int64) ([]*Permission, error)
		FindByPage(ctx context.Context, page, size uint64, keyword string) ([]*Permission, int64, error)
		FindByKeys(ctx context.Context, keys []string) ([]*Permission, error)
		FindAll(ctx context.Context) ([]*Permission, error)
		FindPermissionCodeByUserId(ctx context.Context, userId int64, code string) ([]string, error)
		GetRolePermissionMapsCache(ctx context.Context, rdb *redis.Redis) (map[int64][]string, error)
		RemoveRolePermissionMapsCache(rdb *redis.Redis) error
		WithSession(session sqlx.Session) PermissionModel
	}

	customPermissionModel struct {
		*defaultPermissionModel
	}
)

// NewPermissionModel returns a model for the database table.
func NewPermissionModel(conn sqlx.SqlConn) PermissionModel {
	return &customPermissionModel{
		defaultPermissionModel: newPermissionModel(conn),
	}
}

func (m *customPermissionModel) WithSession(session sqlx.Session) PermissionModel {
	return NewPermissionModel(sqlx.NewSqlConnFromSession(session))
}

func (m *customPermissionModel) CheckPermissionCodeExists(ctx context.Context, code string, excludeId int64) (bool, error) {
	var query string
	var count int64
	var err error

	if excludeId > 0 {
		// For update: check if code exists for other records (excluding current record)
		query = fmt.Sprintf("select count(0) from %s where `code` = ? and `id` != ? limit 1", m.table)
		err = m.conn.QueryRowCtx(ctx, &count, query, code, excludeId)
	} else {
		// For create: check if code exists at all
		query = fmt.Sprintf("select count(0) from %s where `code` = ? limit 1", m.table)
		err = m.conn.QueryRowCtx(ctx, &count, query, code)
	}

	return count > 0, err
}

func (m *customPermissionModel) FindByRoleIds(ctx context.Context, roleIds []int64) ([]*Permission, error) {
	joinRoleIds := make([]string, len(roleIds))
	for i, roleId := range roleIds {
		joinRoleIds[i] = fmt.Sprintf("%d", roleId)
	}
	query := fmt.Sprintf("select p.* from %s p inner join role_permission rp on p.`unique_key` = rp.`permission_key` where rp.`role_id` in (%s)", m.table, strings.Join(joinRoleIds, ","))
	var resp []*Permission
	err := m.conn.QueryRowsCtx(ctx, &resp, query)
	switch err {
	case nil:
		return resp, nil
	case sqlx.ErrNotFound:
		return nil, ErrNotFound
	default:
		return nil, err
	}
}

func (m *customPermissionModel) FindByPage(ctx context.Context, page, size uint64, keyword string) ([]*Permission, int64, error) {
	var resp []*Permission
	var total int64

	whereClause := " WHERE 1=1"
	orderBy := "id ASC"
	var args []interface{}
	if keyword != "" {
		whereClause += " AND name LIKE ?"
		args = append(args, "%"+keyword+"%")
	}
	query := fmt.Sprintf("SELECT COUNT(0) FROM %s %s", m.table, whereClause)
	err := m.conn.QueryRowCtx(ctx, &total, query, args...)
	if err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * size
	args = append(args, offset, size)
	query = fmt.Sprintf("SELECT %s FROM %s %s ORDER BY %s LIMIT ?,?", permissionRows, m.table, whereClause, orderBy)
	err = m.conn.QueryRowsCtx(ctx, &resp, query, args...)
	if err != nil {
		return nil, 0, err
	}

	return resp, total, err
}

func (m *customPermissionModel) FindByKeys(ctx context.Context, keys []string) ([]*Permission, error) {
	if len(keys) == 0 {
		return []*Permission{}, nil
	}

	query := fmt.Sprintf("select %s from %s where `unique_key` in ('%s')", permissionRows, m.table, strings.Join(keys, "','"))
	var resp []*Permission
	err := m.conn.QueryRowsCtx(ctx, &resp, query)
	switch err {
	case nil:
		return resp, nil
	case sqlx.ErrNotFound:
		return nil, ErrNotFound
	default:
		return nil, err
	}
}

func (m *customPermissionModel) FindAll(ctx context.Context) ([]*Permission, error) {
	var resp []*Permission
	query := fmt.Sprintf("select %s from %s", permissionRows, m.table)
	err := m.conn.QueryRowsCtx(ctx, &resp, query)
	switch err {
	case nil:
		return resp, nil
	case sqlx.ErrNotFound:
		return nil, ErrNotFound
	default:
		return nil, err
	}
}

func (m *customPermissionModel) FindPermissionCodeByUserId(ctx context.Context, userId int64, code string) ([]string, error) {
	where := ""
	args := []interface{}{userId}
	if code != "" {
		where = " AND p.code LIKE ?"
		args = append(args, code+"%")
	}
	query := fmt.Sprintf(`
	SELECT DISTINCT p.code FROM %s p INNER JOIN role_permission rp ON p.unique_key = rp.permission_key WHERE rp.role_id IN (
		SELECT role_id FROM user_role WHERE user_id=?
	)%s`, m.table, where)
	var resp []string
	err := m.conn.QueryRowsCtx(ctx, &resp, query, args...)
	switch err {
	case nil:
		return resp, nil
	case sqlx.ErrNotFound:
		return make([]string, 0), ErrNotFound
	default:
		return nil, err
	}
}

// GetRolePermissionMapsCache gets role permission mapping with caching
func (m *customPermissionModel) GetRolePermissionMapsCache(ctx context.Context, rdb *redis.Redis) (map[int64][]string, error) {
	key := "knowdata_role_permission:all"
	rolePermissionMap, err := m.getRolePermissionFromCache(rdb, key)
	if err != nil {
		// Cache miss, get from database
		query := fmt.Sprintf(`
		SELECT r.id as role_id, p.code FROM %s p
		INNER JOIN role_permission rp ON p.unique_key = rp.permission_key
		INNER JOIN role r ON rp.role_id = r.id
		ORDER BY r.id, p.code`, m.table)

		var rolePermissions []RolePermissionData
		err := m.conn.QueryRowsCtx(ctx, &rolePermissions, query)
		if err != nil {
			return nil, err
		}

		// Build role permission map
		rolePermissionMap = make(map[int64][]string)
		for _, rp := range rolePermissions {
			rolePermissionMap[rp.RoleId] = append(rolePermissionMap[rp.RoleId], rp.Code)
		}

		// Update cache in background
		go func() {
			_ = m.saveRolePermissionToCache(rdb, key, rolePermissionMap)
		}()
	}

	return rolePermissionMap, nil
}

// RemoveRolePermissionMapsCache removes the role permission cache
func (m *customPermissionModel) RemoveRolePermissionMapsCache(rdb *redis.Redis) error {
	key := "knowdata_role_permission:all"
	_, err := rdb.Del(key)
	return err
}

// saveRolePermissionToCache saves role permission mapping to redis using msgpack encoding
func (m *customPermissionModel) saveRolePermissionToCache(rdb *redis.Redis, key string, rolePermissionMap map[int64][]string) error {
	data, err := msgpack.Marshal(rolePermissionMap)
	if err != nil {
		return err
	}
	// Cache for 1 hour (3600 seconds)
	err = rdb.Setex(key, string(data), 3600)
	return err
}

// getRolePermissionFromCache retrieves role permission mapping from redis using msgpack encoding
func (m *customPermissionModel) getRolePermissionFromCache(rdb *redis.Redis, key string) (map[int64][]string, error) {
	data, err := rdb.Get(key)
	if err != nil {
		return nil, err
	}

	var rolePermissionMap map[int64][]string
	err = msgpack.Unmarshal([]byte(data), &rolePermissionMap)
	if err != nil {
		return nil, err
	}

	return rolePermissionMap, nil
}

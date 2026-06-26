package bootstrap

import (
	"context"
	"database/sql"
	"fmt"

	"knowsource/api/internal/config"

	_ "github.com/go-sql-driver/mysql"
	"github.com/zeromicro/go-zero/core/stores/redis"
)

// CheckMySQLPing 使用当前配置检测 MySQL 是否可达。
func CheckMySQLPing(c *config.Config) (ok bool, msg string) {
	if c.MySQL.DataSource == "" {
		return false, "未配置 Mysql.DataSource"
	}
	db, err := sql.Open("mysql", c.MySQL.DataSource)
	if err != nil {
		return false, err.Error()
	}
	defer db.Close()
	if err := db.PingContext(context.Background()); err != nil {
		return false, err.Error()
	}
	return true, "可访问"
}

// CheckRedisPing 使用 CacheRedis 第一项检测 Redis 是否可达。
func CheckRedisPing(c *config.Config) (ok bool, msg string) {
	if len(c.CacheRedis) == 0 {
		return false, "未配置 CacheRedis"
	}
	rc := c.CacheRedis[0]
	if rc.Host == "" {
		return false, "Redis Host 为空"
	}
	rdb, err := redis.NewRedis(redis.RedisConf{
		Host: rc.Host,
		Type: rc.Type,
		Pass: rc.Pass,
	})
	if err != nil {
		return false, fmt.Sprintf("连接失败: %v", err)
	}
	if !rdb.PingCtx(context.Background()) {
		return false, "Ping 失败"
	}
	return true, "可访问"
}

// CheckCore 同时检测 MySQL 与 Redis。
func CheckCore(c *config.Config) (mysqlOk, redisOk bool, mysqlMsg, redisMsg string) {
	mysqlOk, mysqlMsg = CheckMySQLPing(c)
	redisOk, redisMsg = CheckRedisPing(c)
	return mysqlOk, redisOk, mysqlMsg, redisMsg
}

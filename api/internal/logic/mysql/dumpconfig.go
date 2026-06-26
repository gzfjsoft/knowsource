package mysql

import (
	"fmt"
	"net"
	"strconv"
	"strings"

	"knowsource/api/internal/config"

	mysqldriver "github.com/go-sql-driver/mysql"
)

// resolveDumpConfig 仅从 Mysql.DataSource 解析 mysqldump 连接参数；可执行文件为 PATH 中的 mysqldump。
func resolveDumpConfig(cfg *config.Config) (dumpBin, host, user, password, database string, port int, err error) {
	dumpBin = "mysqldump"
	ds := strings.TrimSpace(cfg.MySQL.DataSource)
	if ds == "" {
		err = fmt.Errorf("Mysql.DataSource 未配置")
		return
	}
	d, e := mysqldriver.ParseDSN(ds)
	if e != nil {
		err = fmt.Errorf("解析 MySQL DataSource 失败: %w", e)
		return
	}
	host = ""
	if d.Addr != "" {
		h, p, splitErr := net.SplitHostPort(d.Addr)
		if splitErr != nil {
			host = d.Addr
		} else {
			host = h
			port, _ = strconv.Atoi(p)
		}
	}
	user = d.User
	password = d.Passwd
	database = d.DBName
	if port == 0 {
		port = 3306
	}
	if host == "" {
		host = "127.0.0.1"
	}
	if user == "" || database == "" {
		err = fmt.Errorf("Mysql.DataSource 中缺少用户名或数据库名")
		return
	}
	return
}

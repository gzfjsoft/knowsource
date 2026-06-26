package model

import (
	"context"
	"fmt"
	"strings"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ UserAuditLogModel = (*customUserAuditLogModel)(nil)

type (
	// UserAuditLogModel is an interface to be customized, add more methods here,
	// and implement the added methods in customUserAuditLogModel.
	UserAuditLogModel interface {
		userAuditLogModel
		WithSession(session sqlx.Session) UserAuditLogModel
		FindListWithPage(ctx context.Context, page, pageSize uint64, kindKey, kindPrefix string, userId int64, username, keyword string) ([]*UserAuditLog, int64, error)
	}

	customUserAuditLogModel struct {
		*defaultUserAuditLogModel
	}
)

// NewUserAuditLogModel returns a model for the database table.
func NewUserAuditLogModel(conn sqlx.SqlConn) UserAuditLogModel {
	return &customUserAuditLogModel{
		defaultUserAuditLogModel: newUserAuditLogModel(conn),
	}
}

func (m *customUserAuditLogModel) WithSession(session sqlx.Session) UserAuditLogModel {
	return NewUserAuditLogModel(sqlx.NewSqlConnFromSession(session))
}

// FindListWithPage 分页查询用户审计日志列表
func (m *customUserAuditLogModel) FindListWithPage(ctx context.Context, page, pageSize uint64, kindKey, kindPrefix string, userId int64, username, keyword string) ([]*UserAuditLog, int64, error) {
	// 构建查询条件
	var conditions []string
	var args []interface{}

	// 日志类型筛选
	if kindKey != "" {
		conditions = append(conditions, "`kind_key` = ?")
		args = append(args, kindKey)
	} else if kindPrefix != "" {
		conditions = append(conditions, "`kind_key` LIKE ?")
		args = append(args, kindPrefix+"_%")
	}

	// 用户ID筛选
	if userId > 0 {
		conditions = append(conditions, "`user_id` = ?")
		args = append(args, userId)
	} else if username != "" {
		conditions = append(conditions, "`username` LIKE ?")
		args = append(args, "%"+username+"%")
	}

	// 关键词搜索（搜索来源标题和备注）
	if keyword != "" {
		conditions = append(conditions, "(`source_title` LIKE ? OR `remark` LIKE ?)")
		args = append(args, "%"+keyword+"%", "%"+keyword+"%")
	}

	// 构建WHERE子句
	whereClause := ""
	if len(conditions) > 0 {
		whereClause = "WHERE " + strings.Join(conditions, " AND ")
	}

	// 查询总数
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM %s %s", m.table, whereClause)
	var total int64
	err := m.conn.QueryRowCtx(ctx, &total, countQuery, args...)
	if err != nil {
		return nil, 0, err
	}

	// 查询分页数据
	offset := (page - 1) * pageSize
	dataQuery := fmt.Sprintf(`
		SELECT %s FROM %s %s
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?`, userAuditLogRows, m.table, whereClause)

	// 添加分页参数到args
	queryArgs := append(args, pageSize, offset)

	var auditLogs []*UserAuditLog
	err = m.conn.QueryRowsCtx(ctx, &auditLogs, dataQuery, queryArgs...)
	if err != nil && err != sqlx.ErrNotFound {
		return nil, 0, err
	}

	return auditLogs, total, nil
}

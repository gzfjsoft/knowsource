package model

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ MinuteUsageModel = (*customMinuteUsageModel)(nil)

type (
	// MinuteUsageModel is an interface to be customized, add more methods here,
	// and implement the added methods in customMinuteUsageModel.
	MinuteUsageModel interface {
		minuteUsageModel
		WithSession(session sqlx.Session) MinuteUsageModel
		FindOneByRunresIdDaynum(ctx context.Context, runresId uint64, daynum uint64) (*MinuteUsage, error)
		List(ctx context.Context, usageId, orgId, userId *uint64, startDatetime, endDatetime *uint64, page, pageSize uint64) ([]*MinuteUsageView, uint64, error)
		Trans(ctx context.Context, fn func(ctx context.Context, session sqlx.Session) error) error
		FindLastRecord(ctx context.Context) (*MinuteUsage, error)
		FindByTimeRange(ctx context.Context, startTime, endTime time.Time) ([]*MinuteUsage, error)
	}

	customMinuteUsageModel struct {
		*defaultMinuteUsageModel
	}
)

// NewMinuteUsageModel returns a model for the database table.
func NewMinuteUsageModel(conn sqlx.SqlConn) MinuteUsageModel {
	return &customMinuteUsageModel{
		defaultMinuteUsageModel: newMinuteUsageModel(conn),
	}
}

func (m *customMinuteUsageModel) WithSession(session sqlx.Session) MinuteUsageModel {
	return NewMinuteUsageModel(sqlx.NewSqlConnFromSession(session))
}

type MinuteUsageView struct {
	MinuteUsage
	InstanceName string `db:"instance_name"`
	ResourceName string `db:"resource_name"`
}

func (m *customMinuteUsageModel) FindOneByRunresIdDaynum(ctx context.Context, runresId uint64, daynum uint64) (*MinuteUsage, error) {
	var resp MinuteUsage
	query := fmt.Sprintf("select %s from %s where `runres_id` = ? and `daynum` = ? limit 1", minuteUsageRows, m.table)
	err := m.conn.QueryRowCtx(ctx, &resp, query, runresId, daynum)
	switch err {
	case nil:
		return &resp, nil
	case sqlx.ErrNotFound:
		return nil, ErrNotFound
	default:
		return nil, err
	}
}

func (m *customMinuteUsageModel) List(ctx context.Context, usageId, orgId, userId *uint64, startDatetime, endDatetime *uint64, page, pageSize uint64) ([]*MinuteUsageView, uint64, error) {
	conditions := []string{"1=1"}
	args := []interface{}{}

	if usageId != nil {
		conditions = append(conditions, "a.`usage_id` = ?")
		args = append(args, *usageId)
	}
	if orgId != nil {
		conditions = append(conditions, "a.`org_id` = ?")
		args = append(args, *orgId)
	}
	if userId != nil {
		conditions = append(conditions, "a.`user_id` = ?")
		args = append(args, *userId)
	}
	if startDatetime != nil {
		conditions = append(conditions, "a.`usage_datetime` >= ?")
		args = append(args, *startDatetime)
	}
	if endDatetime != nil {
		conditions = append(conditions, "a.`usage_datetime` <= ?")
		args = append(args, *endDatetime)
	}

	whereClause := strings.Join(conditions, " AND ")

	whereClauseCount := strings.Replace(whereClause, "a.`", "`", -1)

	// Get total count
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE %s", m.table, whereClauseCount)
	var total uint64
	err := m.conn.QueryRowCtx(ctx, &total, countQuery, args...)
	if err != nil {
		return nil, 0, err
	}

	// Get paginated data
	offset := (page - 1) * pageSize
	fields := []string{}
	for _, field := range minuteUsageFieldNames {
		fields = append(fields, fmt.Sprintf("a.%s", field))
	}
	allfields := strings.Join(fields, ",")

	query := fmt.Sprintf(`SELECT %s, ifnull(instances.name,"") as instance_name, ifnull(running_resources.resource_name,"") as resource_name 
	 FROM %s a 
LEFT JOIN 
    instances ON a.instance_id = instances.instance_id 
LEFT JOIN 
    running_resources ON a.runres_id = running_resources.runres_id  
WHERE %s ORDER BY a.usage_datetime DESC LIMIT ? OFFSET ?`, allfields, m.table, whereClause)
	args = append(args, pageSize, offset)

	var resp []*MinuteUsageView
	err = m.conn.QueryRowsCtx(ctx, &resp, query, args...)
	if err != nil {
		return nil, 0, err
	}

	return resp, total, nil
}

func (m *customMinuteUsageModel) Trans(ctx context.Context, fn func(ctx context.Context, session sqlx.Session) error) error {
	return m.conn.TransactCtx(ctx, fn)
}

func (m *customMinuteUsageModel) FindLastRecord(ctx context.Context) (*MinuteUsage, error) {
	query := fmt.Sprintf("select %s from %s order by id desc limit 1", minuteUsageRows, m.table)
	var resp MinuteUsage
	err := m.conn.QueryRowCtx(ctx, &resp, query)
	switch err {
	case nil:
		return &resp, nil
	case sqlx.ErrNotFound:
		return nil, ErrNotFound
	default:
		return nil, err
	}
}

func (m *customMinuteUsageModel) FindByTimeRange(ctx context.Context, startTime, endTime time.Time) ([]*MinuteUsage, error) {
	query := fmt.Sprintf("select %s from %s where usage_datetime between ? and ?", minuteUsageRows, m.table)
	var resp []*MinuteUsage
	err := m.conn.QueryRowsCtx(ctx, &resp, query, startTime, endTime)
	return resp, err
}

package model

import (
	"context"
	"fmt"
	"strings"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ DailyUsageModel = (*customDailyUsageModel)(nil)

type (
	// DailyUsageModel is an interface to be customized, add more methods here,
	// and implement the added methods in customDailyUsageModel.
	DailyUsageModel interface {
		dailyUsageModel
		withSession(session sqlx.Session) DailyUsageModel
		List(ctx context.Context, orgId, userId *uint64, startDate, endDate *uint64, page, pageSize uint64) ([]*DailyUsageView, uint64, error)
		FindByDateAndResource(ctx context.Context, orgId, userId, runres_id, instanceId uint64, daynum uint64) (*DailyUsage, error)
		FindLatestByInstanceId(ctx context.Context, orgId uint64, userId uint64, instanceId uint64) (*DailyUsage, error)
	}

	customDailyUsageModel struct {
		*defaultDailyUsageModel
	}
)

// NewDailyUsageModel returns a model for the database table.
func NewDailyUsageModel(conn sqlx.SqlConn) DailyUsageModel {
	return &customDailyUsageModel{
		defaultDailyUsageModel: newDailyUsageModel(conn),
	}
}

func (m *customDailyUsageModel) withSession(session sqlx.Session) DailyUsageModel {
	return NewDailyUsageModel(sqlx.NewSqlConnFromSession(session))
}

type DailyUsageView struct {
	DailyUsage
	InstanceName string `db:"instance_name"`
	ResourceName string `db:"resource_name"`
}

func (m *customDailyUsageModel) FindLatestByInstanceId(ctx context.Context, orgId uint64, userId uint64, instanceId uint64) (*DailyUsage, error) {
	query := fmt.Sprintf("select * from %s where org_id = ? and user_id = ? and instance_id = ?  order by daynum desc limit 1", m.table)
	var resp DailyUsage
	err := m.conn.QueryRowCtx(ctx, &resp, query, orgId, userId, instanceId)
	return &resp, err
}

func (m *customDailyUsageModel) List(ctx context.Context, orgId, userId *uint64, startDate, endDate *uint64, page, pageSize uint64) ([]*DailyUsageView, uint64, error) {
	conditions := []string{" 1=1 "}
	args := []interface{}{}

	if orgId != nil {
		conditions = append(conditions, "du.org_id = ?")
		args = append(args, *orgId)
	}
	if userId != nil {
		conditions = append(conditions, "du.user_id = ?")
		args = append(args, *userId)
	}
	if startDate != nil {
		conditions = append(conditions, "du.usage_date >= ?")
		args = append(args, *startDate)
	}
	if endDate != nil {
		conditions = append(conditions, "du.usage_date <= ?")
		args = append(args, *endDate)
	}

	whereClause := strings.Join(conditions, " AND ")

	// Get total count
	countQuery := fmt.Sprintf(`
		SELECT COUNT(*) 
		FROM %s du 
		WHERE %s`, m.table, whereClause)
	var total uint64
	err := m.conn.QueryRowCtx(ctx, &total, countQuery, args...)
	if err != nil {
		return nil, 0, err
	}

	// Get paginated data with joined tables
	offset := (page - 1) * pageSize
	query := fmt.Sprintf(`
		SELECT du.*, ifnull(ri.name,"") as instance_name, ifnull(rr.resource_name,"") as resource_name 
 		FROM %s du 
		LEFT JOIN  instances ri ON du.instance_id = ri.instance_id 
		LEFT JOIN running_resources rr ON du.runres_id = rr.runres_id 
		WHERE %s 
		ORDER BY du.usage_date DESC 
		LIMIT ? OFFSET ?`, m.table, whereClause)
	args = append(args, pageSize, offset)

	var resp []*DailyUsageView
	err = m.conn.QueryRowsCtx(ctx, &resp, query, args...)
	if err != nil {
		return nil, 0, err
	}

	return resp, total, nil
}

func (m *customDailyUsageModel) FindByDateAndResource(ctx context.Context, orgId, userId, runres_id, instanceId uint64, daynum uint64) (*DailyUsage, error) {
	query := fmt.Sprintf("select %s from %s where org_id = ? and user_id = ? and runres_id = ? and instance_id = ? and daynum = ? limit 1", dailyUsageRows, m.table)
	var resp DailyUsage
	err := m.conn.QueryRowCtx(ctx, &resp, query, orgId, userId, runres_id, instanceId, daynum)
	switch err {
	case nil:
		return &resp, nil
	case sqlx.ErrNotFound:
		return nil, ErrNotFound
	default:
		return nil, err
	}
}

package model

import (
	"context"
	"fmt"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ ResourceOrgsModel = (*customResourceOrgsModel)(nil)

type (
	// ResourceOrgsModel is an interface to be customized, add more methods here,
	// and implement the added methods in customResourceOrgsModel.
	ResourceOrgsModel interface {
		resourceOrgsModel
		withSession(session sqlx.Session) ResourceOrgsModel
		DeleteByUserIDOrgID(ctx context.Context, rid uint64, oid uint64) error
		FindByResourceIdOrgId(ctx context.Context, resourceId, orgId uint64) (*ResourceOrgs, error)
		FindByResourceId(ctx context.Context, resourceId uint64) ([]*ResourceOrgs, error)
		FindByOrgId(ctx context.Context, orgId uint64) ([]*ResourceOrgs, error)
		FindAll(ctx context.Context) ([]*ResourceOrgs, error)
		FindDiscountByResourceIdOrgId(ctx context.Context, resourceId, orgId uint64) (*ResourceOrgDiscount, error)
	}

	customResourceOrgsModel struct {
		*defaultResourceOrgsModel
	}
)

// NewResourceOrgsModel returns a model for the database table.
func NewResourceOrgsModel(conn sqlx.SqlConn) ResourceOrgsModel {
	return &customResourceOrgsModel{
		defaultResourceOrgsModel: newResourceOrgsModel(conn),
	}
}

func (m *customResourceOrgsModel) withSession(session sqlx.Session) ResourceOrgsModel {
	return NewResourceOrgsModel(sqlx.NewSqlConnFromSession(session))
}

func (m *customResourceOrgsModel) DeleteByUserIDOrgID(ctx context.Context, rid uint64, oid uint64) error {
	query := fmt.Sprintf("delete from %s where `resource_id` = ? and  `org_id` = ?", m.table)
	_, err := m.conn.ExecCtx(ctx, query, rid, oid)
	return err
}

type ResourceOrgDiscount struct {
	Id               uint64 `db:"id"`
	OrgId            uint64 `db:"org_id"`
	ResourceId       uint64 `db:"resource_id"`
	DiscountId       uint64 `db:"discount_id"`
	CreatedAt        uint64 `db:"created_at"`
	OrgName          string `db:"org_name"`
	ResourceName     string `db:"resource_name"`
	ResourceType     string `db:"resource_type"`
	UnitHourlyPrice  int64  `db:"unit_hourly_price"`
	UnitDailyPrice   int64  `db:"unit_daily_price"`
	UnitMonthlyPrice int64  `db:"unit_monthly_price"`
	UnitYearlyPrice  int64  `db:"unit_yearly_price"`
	Total            int64  `db:"total"`
	Remains          int64  `db:"remains"`
	Memo             string `db:"memo"`
	HourlyDiscount   int64  `db:"hourly_discount"`
	DailyDiscount    int64  `db:"daily_discount"`
	MonthlyDiscount  int64  `db:"monthly_discount"`
	YearlyDiscount   int64  `db:"yearly_discount"`
}

func (m *customResourceOrgsModel) FindDiscountByResourceIdOrgId(ctx context.Context, resourceId, orgId uint64) (*ResourceOrgDiscount, error) {
	sql := `SELECT
				ro.id,
				ro.org_id,
				ro.resource_id,
				ro.discount_id,
				UNIX_TIMESTAMP(ro.created_at) as created_at,
				o.org_name,
				r.resource_name,
				r.resource_type,
				r.unit_hourly_price,
				r.unit_daily_price,
				r.unit_monthly_price,
				r.unit_yearly_price,
				r.total,
				r.remains,
				IFNULL( d.memo,"") as memo,
				IFNULL(d.min_discount,0) min_discount,
				IFNULL(d.hourly_discount,0) hourly_discount,
				IFNULL(d.daily_discount,0) daily_discount,
				IFNULL(d.monthly_discount,0) monthly_discount,
				IFNULL(d.yearly_discount,0) yearly_discount

			FROM
				resource_orgs ro
			LEFT JOIN resources r ON ro.resource_id = r.resource_id
			LEFT JOIN organizations o ON ro.org_id = o.org_id
			LEFT JOIN resource_discounts d ON ro.discount_id = d.discount_id and ro.org_id= d.org_id
			WHERE
				ro.resource_id = ? and ro.org_id = ?`

	var resourceOrgDiscount ResourceOrgDiscount

	err := m.conn.QueryRowCtx(ctx, &resourceOrgDiscount, sql, resourceId, orgId)
	if err != nil {
		return nil, err
	}
	return &resourceOrgDiscount, nil
}

func (m *customResourceOrgsModel) FindByResourceIdOrgId(ctx context.Context, resourceId, orgId uint64) (*ResourceOrgs, error) {
	query := fmt.Sprintf("select %s from %s where resource_id = ? and org_id = ? limit 1", resourceOrgsRows, m.table)
	var resp ResourceOrgs
	err := m.conn.QueryRowCtx(ctx, &resp, query, resourceId, orgId)
	switch err {
	case nil:
		return &resp, nil
	case sqlx.ErrNotFound:
		return nil, ErrNotFound
	default:
		return nil, err
	}
}

func (m *customResourceOrgsModel) FindByResourceId(ctx context.Context, resourceId uint64) ([]*ResourceOrgs, error) {
	query := fmt.Sprintf("select %s from %s where resource_id = ?", resourceOrgsRows, m.table)
	var resp []*ResourceOrgs
	err := m.conn.QueryRowsCtx(ctx, &resp, query, resourceId)
	return resp, err
}

func (m *customResourceOrgsModel) FindByOrgId(ctx context.Context, orgId uint64) ([]*ResourceOrgs, error) {
	query := fmt.Sprintf("select %s from %s where org_id = ?", resourceOrgsRows, m.table)
	var resp []*ResourceOrgs
	err := m.conn.QueryRowsCtx(ctx, &resp, query, orgId)
	return resp, err
}

func (m *customResourceOrgsModel) FindAll(ctx context.Context) ([]*ResourceOrgs, error) {
	query := fmt.Sprintf("select %s from %s", resourceOrgsRows, m.table)
	var resp []*ResourceOrgs
	err := m.conn.QueryRowsCtx(ctx, &resp, query)
	return resp, err
}

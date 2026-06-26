package model

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ BalancesModel = (*customBalancesModel)(nil)

type (
	// BalancesModel is an interface to be customized, add more methods here,
	// and implement the added methods in customBalancesModel.
	BalancesModel interface {
		balancesModel
		WithSession(session sqlx.Session) BalancesModel
		FindList(ctx context.Context, condition string) (*[]Balances, error)
		FindListView(ctx context.Context, condition string) (*[]UserBalance, error)
		FindOneByUserAndCurrency(ctx context.Context, userId uint64, OrgId uint64, currencyCode string) (*Balances, error)
		UpdateBalance(ctx context.Context, data *Balances) error

		AdjBalance(ctx context.Context, userId uint64, orgId uint64, currencyCode string, money int64) error

		DeleteByUserId(ctx context.Context, userId uint64) error
		DeleteByUserIdAndOrgId(ctx context.Context, userId uint64, orgId uint64) error
		Charging(ctx context.Context, userId uint64, orgId uint64, money int64, orderNo string, currencyCode string, detail string) (*Balances, error)
	}

	customBalancesModel struct {
		*defaultBalancesModel
	}
)

// NewBalancesModel returns a model for the database table.
func NewBalancesModel(conn sqlx.SqlConn) BalancesModel {
	return &customBalancesModel{
		defaultBalancesModel: newBalancesModel(conn),
	}
}

func (m *customBalancesModel) AdjBalance(ctx context.Context, userId uint64, orgId uint64, currencyCode string, money int64) error {
	query := fmt.Sprintf("update %s set balance = balance + ? where `user_id` = ? and `org_id` = ? and `currency_code` = ?", m.table)
	_, err := m.conn.ExecCtx(ctx, query, money, userId, orgId, currencyCode)
	return err
}

func (m *customBalancesModel) WithSession(session sqlx.Session) BalancesModel {
	return NewBalancesModel(sqlx.NewSqlConnFromSession(session))
}

func (m *customBalancesModel) DeleteByUserId(ctx context.Context, userId uint64) error {
	query := fmt.Sprintf("delete from %s where `user_id` = ?", m.table)
	_, err := m.conn.ExecCtx(ctx, query, userId)
	return err
}

func (m *customBalancesModel) FindOneByUserAndCurrency(ctx context.Context, userId uint64, OrgId uint64, currencyCode string) (*Balances, error) {
	var resp Balances
	query := fmt.Sprintf("select %s from %s where `user_id` = ? and `currency_code` = ? and org_id = ? limit 1", balancesRows, m.table)
	err := m.conn.QueryRowCtx(ctx, &resp, query, userId, currencyCode, OrgId)
	switch err {
	case nil:
		return &resp, nil
	case sqlx.ErrNotFound:
		return nil, ErrNotFound
	default:
		return nil, err
	}
}

func (m *customBalancesModel) UpdateBalance(ctx context.Context, data *Balances) error {
	query := fmt.Sprintf("update %s set balance=?, updated_at=? where `user_id` = ? and `currency_code` = ? and `org_id` = ?", m.table)
	_, err := m.conn.ExecCtx(ctx, query, data.Balance, time.Now(), data.UserId, data.CurrencyCode, data.OrgId)
	return err
}

type UserBalance struct {
	UserId       uint64 `db:"user_id"`
	OrgId        uint64 `db:"org_id"`
	OrgName      string `db:"org_name"`
	CurrencyCode string `db:"currency_code"`
	Balance      int64  `db:"balance"`
}

func (m *customBalancesModel) FindListView(ctx context.Context, condition string) (*[]UserBalance, error) {
	query := fmt.Sprintf("select balances.user_id, balances.org_id,organizations.org_name,balances.balance,balances.currency_code from balances left outer join organizations on organizations.org_id = balances.org_id %s", condition)
	var resp []UserBalance
	err := m.conn.QueryRowsCtx(ctx, &resp, query)
	switch err {
	case nil:
		return &resp, nil
	case sqlx.ErrNotFound:
		return nil, ErrNotFound
	default:
		return nil, err
	}
}

func (m *customBalancesModel) FindList(ctx context.Context, condition string) (*[]Balances, error) {
	query := fmt.Sprintf("select %s from %s %s", balancesRows, m.table, condition)
	var resp []Balances
	err := m.conn.QueryRowsCtx(ctx, &resp, query)
	switch err {
	case nil:
		return &resp, nil
	case sqlx.ErrNotFound:
		return nil, ErrNotFound
	default:
		return nil, err
	}
}

// Charging handles balance charging with transaction
func (m *customBalancesModel) Charging(ctx context.Context, userId uint64, orgId uint64, money int64, orderNo string, currencyCode string, detail string) (*Balances, error) {
	var balance *Balances

	err := m.conn.TransactCtx(ctx, func(ctx context.Context, session sqlx.Session) error {
		balModel := m.WithSession(session)

		// Find or initialize balance record
		item, err := balModel.FindOneByUserAndCurrency(ctx, userId, orgId, currencyCode)
		if err == ErrNotFound {
			item = &Balances{
				UserId:       userId,
				OrgId:        orgId,
				Balance:      0,
				CurrencyCode: currencyCode,
			}
		} else if err != nil {
			return err
		}

		// Update balance
		item.Balance -= money
		if item.BalanceId > 0 {
			err = balModel.UpdateBalance(ctx, item)
		} else {
			_, err = balModel.Insert(ctx, item)
		}
		if err != nil {
			return err
		}

		// Create transaction record
		transRecord := &TransactionRecords{
			TransType:    2, // 1-入金; 2-出金
			OrgId:        orgId,
			CurrencyCode: currencyCode,
			UserId:       userId,
			PayType:      "",
			Detail:       detail,
			Amount:       money,
			OrderNo:      sql.NullString{String: orderNo, Valid: true},
			Balance:      item.Balance,
		}

		// Insert transaction record using the same session
		_, err = NewTransactionRecordsModel(sqlx.NewSqlConnFromSession(session)).Insert(ctx, transRecord)
		if err != nil {
			return err
		}

		balance = item
		return nil
	})

	if err != nil {
		msg := fmt.Sprintf("Charging 失败!!! %v,(uid:%d,orgid:%d,money:%d,orderNo:%s,currencyCode:%s,detail:%s)", err, userId, orgId, money, orderNo, currencyCode, detail)
		logx.Errorf(msg)

		NewErrorLogModel(m.conn).Insert(ctx, &ErrorLog{
			Message: msg,
			Tag:     "fee",
		})
		return nil, err
	}

	return balance, nil
}

func (m *customBalancesModel) DeleteByUserIdAndOrgId(ctx context.Context, userId uint64, orgId uint64) error {
	query := fmt.Sprintf("delete from %s where `user_id` = ? and `org_id` = ?", m.table)
	_, err := m.conn.ExecCtx(ctx, query, userId, orgId)
	return err
}

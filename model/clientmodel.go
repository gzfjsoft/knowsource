package model

import (
	"context"
	"fmt"
	"strings"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ ClientModel = (*customClientModel)(nil)

type (
	// ClientModel is an interface to be customized, add more methods here,
// and implement the added methods in customClientModel.
ClientModel interface {
	clientModel
	withSession(session sqlx.Session) ClientModel
	List(ctx context.Context, clientIdLike string, page, pageSize uint64) ([]*Client, int64, error)
	DeleteByClientId(ctx context.Context, clientId string) error
	FindAll(ctx context.Context) ([]*Client, error)
	FindOneByOwnerEmail(ctx context.Context, ownerEmail string) (*Client, error)
}

	customClientModel struct {
		*defaultClientModel
	}
)

// NewClientModel returns a model for the database table.
func NewClientModel(conn sqlx.SqlConn) ClientModel {
	return &customClientModel{
		defaultClientModel: newClientModel(conn),
	}
}

func (m *customClientModel) withSession(session sqlx.Session) ClientModel {
	return NewClientModel(sqlx.NewSqlConnFromSession(session))
}

func (m *customClientModel) Update(ctx context.Context, newData *Client) error {
	query := fmt.Sprintf("UPDATE %s SET `client_json_info`=?, `registration_ip`=?, `status`=?, `verified_at`=?, `owner_email`=? WHERE `client_id`=?", m.table)
	_, err := m.conn.ExecCtx(ctx, query,
		newData.ClientJsonInfo,
		newData.RegistrationIp,
		newData.Status,
		newData.VerifiedAt,
		newData.OwnerEmail,
		newData.ClientId,
	)
	return err
}

func (m *customClientModel) FindOneByOwnerEmail(ctx context.Context, ownerEmail string) (*Client, error) {
	ownerEmail = strings.TrimSpace(ownerEmail)
	if ownerEmail == "" {
		return nil, ErrNotFound
	}
	var resp Client
	q := fmt.Sprintf("SELECT %s FROM %s WHERE `owner_email` = ? LIMIT 1", clientRows, m.table)
	err := m.conn.QueryRowCtx(ctx, &resp, q, ownerEmail)
	switch err {
	case nil:
		return &resp, nil
	case sqlx.ErrNotFound:
		return nil, ErrNotFound
	default:
		return nil, err
	}
}

func (m *customClientModel) List(ctx context.Context, clientIdLike string, page, pageSize uint64) ([]*Client, int64, error) {
	where := " where 1=1 "
	args := make([]interface{}, 0, 2)
	if strings.TrimSpace(clientIdLike) != "" {
		where += " and client_id like ? "
		args = append(args, "%"+strings.TrimSpace(clientIdLike)+"%")
	}

	var total int64
	countQuery := fmt.Sprintf("select count(1) from %s %s", m.table, where)
	if err := m.conn.QueryRowCtx(ctx, &total, countQuery, args...); err != nil {
		return nil, 0, err
	}

	if page == 0 {
		page = 1
	}
	if pageSize == 0 {
		pageSize = 10
	}
	offset := (page - 1) * pageSize
	query := fmt.Sprintf("select %s from %s %s order by updated_at desc limit %d, %d", clientRows, m.table, where, offset, pageSize)
	var list []*Client
	err := m.conn.QueryRowsCtx(ctx, &list, query, args...)
	return list, total, err
}

func (m *customClientModel) DeleteByClientId(ctx context.Context, clientId string) error {
	query := fmt.Sprintf("delete from %s where `client_id` = ?", m.table)
	_, err := m.conn.ExecCtx(ctx, query, clientId)
	return err
}

func (m *customClientModel) FindAll(ctx context.Context) ([]*Client, error) {
	query := fmt.Sprintf("select %s from %s", clientRows, m.table)
	var list []*Client
	err := m.conn.QueryRowsCtx(ctx, &list, query)
	return list, err
}

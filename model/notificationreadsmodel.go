package model

import (
	"context"
	"fmt"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ NotificationReadsModel = (*customNotificationReadsModel)(nil)

type (
	// NotificationReadsModel is an interface to be customized, add more methods here,
	// and implement the added methods in customNotificationReadsModel.
	NotificationReadsModel interface {
		notificationReadsModel
		WithSession(session sqlx.Session) NotificationReadsModel
		FindAllByFilter(ctx context.Context, filter string) ([]*NotificationReads, error)
	}

	customNotificationReadsModel struct {
		*defaultNotificationReadsModel
	}
)

// NewNotificationReadsModel returns a model for the database table.
func NewNotificationReadsModel(conn sqlx.SqlConn) NotificationReadsModel {
	return &customNotificationReadsModel{
		defaultNotificationReadsModel: newNotificationReadsModel(conn),
	}
}

func (m *customNotificationReadsModel) WithSession(session sqlx.Session) NotificationReadsModel {
	return NewNotificationReadsModel(sqlx.NewSqlConnFromSession(session))
}

func (m *customNotificationReadsModel) FindAllByFilter(ctx context.Context, filter string) ([]*NotificationReads, error) {
	query := fmt.Sprintf("select * from %s where %s", m.table, filter)
	var resp []*NotificationReads

	err := m.defaultNotificationReadsModel.conn.QueryRowsCtx(ctx, &resp, query)
	switch err {
	case nil:
		return resp, nil
	case sqlx.ErrNotFound:
		return nil, ErrNotFound
	}
	return resp, nil
}

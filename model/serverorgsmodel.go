package model

import (
	"context"
	"fmt"
	"strconv"

	"github.com/zeromicro/go-zero/core/logc"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ ServerOrgsModel = (*customServerOrgsModel)(nil)

type (
	// ServerOrgsModel is an interface to be customized, add more methods here,
	// and implement the added methods in customServerOrgsModel.
	ServerOrgsModel interface {
		serverOrgsModel
		withSession(session sqlx.Session) ServerOrgsModel
		DeleteByServerOrg(ctx context.Context, serverOrg *ServerOrgs) error
		FindOneByServerOrg(ctx context.Context, orgId uint64, serverId uint64) (*ServerOrgs, error)
		FindAllByServerOrg(ctx context.Context, orgId uint64, serverId uint64) ([]*ServerOrgs, error)
	}

	customServerOrgsModel struct {
		*defaultServerOrgsModel
	}
)

// NewServerOrgsModel returns a model for the database table.
func NewServerOrgsModel(conn sqlx.SqlConn) ServerOrgsModel {
	return &customServerOrgsModel{
		defaultServerOrgsModel: newServerOrgsModel(conn),
	}
}

func (m *customServerOrgsModel) withSession(session sqlx.Session) ServerOrgsModel {
	return NewServerOrgsModel(sqlx.NewSqlConnFromSession(session))
}

func (m *customServerOrgsModel) DeleteByServerOrg(ctx context.Context, serverOrg *ServerOrgs) error {
	query := fmt.Sprintf("delete from %s where `server_id` = ? and `org_id` = ?", m.table)
	_, err := m.conn.ExecCtx(ctx, query, serverOrg.ServerId, serverOrg.OrgId)
	return err
}
func (m *customServerOrgsModel) FindOneByServerOrg(ctx context.Context, orgId uint64, serverId uint64) (*ServerOrgs, error) {
	query := fmt.Sprintf("select %s from %s where `org_id` = ? and `server_id` = ? limit 1", serverOrgsRows, m.table)
	var resp ServerOrgs
	err := m.conn.QueryRowCtx(ctx, &resp, query, orgId, serverId)
	switch err {
	case nil:
		return &resp, nil
	case sqlx.ErrNotFound:
		return nil, ErrNotFound
	default:
		return nil, err
	}
}

func (m *customServerOrgsModel) FindAllByServerOrg(ctx context.Context, orgId uint64, serverId uint64) ([]*ServerOrgs, error) {
	condition := "1=1"
	if orgId != 0 {
		condition = condition + " and org_id = " + strconv.Itoa(int(orgId))
	}

	if serverId != 0 {
		condition = condition + " and server_id = " + strconv.Itoa(int(serverId))
	}

	query := fmt.Sprintf("select %s from %s where %s", serverOrgsRows, m.table, condition)
	logc.Info(context.Background(), query)
	var resp []*ServerOrgs
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

package model

import (
	"context"
	"fmt"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ InvitationModel = (*customInvitationModel)(nil)

type (
	// InvitationModel is an interface to be customized, add more methods here,
	// and implement the added methods in customInvitationModel.
	InvitationModel interface {
		invitationModel
		withSession(session sqlx.Session) InvitationModel
		FindAll(ctx context.Context, uid int64, page, pageSize int64) (*[]Invitation, error)
		Count(ctx context.Context, uid int64) (int64, error)
		FindAllByInviteeId(ctx context.Context, uid int64, page, pageSize int64) (*[]Invitation, error)
		CountByInviteeId(ctx context.Context, uid int64) (int64, error)
		FindOneByToken(ctx context.Context, token string) (*Invitation, error)
	}

	customInvitationModel struct {
		*defaultInvitationModel
	}
)

// NewInvitationModel returns a model for the database table.
func NewInvitationModel(conn sqlx.SqlConn) InvitationModel {
	return &customInvitationModel{
		defaultInvitationModel: newInvitationModel(conn),
	}
}

func (m *customInvitationModel) withSession(session sqlx.Session) InvitationModel {
	return NewInvitationModel(sqlx.NewSqlConnFromSession(session))
}

func (m *defaultInvitationModel) FindAll(ctx context.Context, uid int64, page, pageSize int64) (*[]Invitation, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}
	offset := (page - 1) * pageSize

	query := fmt.Sprintf("select %s from %s where inviter_id = ? limit ?,?", invitationRows, m.table)
	var resp []Invitation
	err := m.conn.QueryRowsCtx(ctx, &resp, query, uid, offset, pageSize)
	switch err {
	case nil:
		return &resp, nil
	case sqlx.ErrNotFound:
		return nil, ErrNotFound
	default:
		return nil, err
	}
}

func (m *defaultInvitationModel) Count(ctx context.Context, uid int64) (int64, error) {
	query := fmt.Sprintf("select count(*) from %s where inviter_id = ?", m.table)
	var count int64
	err := m.conn.QueryRowCtx(ctx, &count, query, uid)
	switch err {
	case nil:
		return count, nil
	default:
		return 0, err
	}
}

func (m *defaultInvitationModel) FindAllByInviteeId(ctx context.Context, uid int64, page, pageSize int64) (*[]Invitation, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}
	offset := (page - 1) * pageSize

	query := fmt.Sprintf("select %s from %s where invitee_id = ? limit ?,?", invitationRows, m.table)
	var resp []Invitation
	err := m.conn.QueryRowsCtx(ctx, &resp, query, uid, offset, pageSize)
	switch err {
	case nil:
		return &resp, nil
	case sqlx.ErrNotFound:
		return nil, ErrNotFound
	default:
		return nil, err
	}
}

func (m *defaultInvitationModel) CountByInviteeId(ctx context.Context, uid int64) (int64, error) {
	query := fmt.Sprintf("select count(*) from %s where invitee_id = ?", m.table)
	var count int64
	err := m.conn.QueryRowCtx(ctx, &count, query, uid)
	switch err {
	case nil:
		return count, nil
	default:
		return 0, err
	}
}

func (m *defaultInvitationModel) FindOneByToken(ctx context.Context, token string) (*Invitation, error) {
	query := fmt.Sprintf("select %s from %s where `invitation_token` = ? limit 1", invitationRows, m.table)
	var resp Invitation
	err := m.conn.QueryRowCtx(ctx, &resp, query, token)
	switch err {
	case nil:
		return &resp, nil
	case sqlx.ErrNotFound:
		return nil, ErrNotFound
	default:
		return nil, err
	}
}

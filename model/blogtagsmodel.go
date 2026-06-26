package model

import (
	"context"
	"fmt"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ BlogTagsModel = (*customBlogTagsModel)(nil)

type (
	// BlogTagsModel is an interface to be customized, add more methods here,
	// and implement the added methods in customBlogTagsModel.
	BlogTagsModel interface {
		blogTagsModel
		withSession(session sqlx.Session) BlogTagsModel
		FindAll(ctx context.Context) ([]*BlogTags, error)
		FindOneByTag(ctx context.Context, tag string) (*BlogTags, error)
	}

	customBlogTagsModel struct {
		*defaultBlogTagsModel
	}
)

// NewBlogTagsModel returns a model for the database table.
func NewBlogTagsModel(conn sqlx.SqlConn) BlogTagsModel {
	return &customBlogTagsModel{
		defaultBlogTagsModel: newBlogTagsModel(conn),
	}
}

func (m *customBlogTagsModel) withSession(session sqlx.Session) BlogTagsModel {
	return NewBlogTagsModel(sqlx.NewSqlConnFromSession(session))
}

func (m *defaultBlogTagsModel) FindAll(ctx context.Context) ([]*BlogTags, error) {
	query := fmt.Sprintf("select %s from %s", blogTagsRows, m.table)
	var resp []*BlogTags
	err := m.conn.QueryRowsCtx(ctx, &resp, query)
	return resp, err
}

func (m *defaultBlogTagsModel) FindOneByTag(ctx context.Context, tag string) (*BlogTags, error) {
	query := fmt.Sprintf("select %s from %s where `tag` = ? limit 1", blogTagsRows, m.table)
	var resp BlogTags
	err := m.conn.QueryRowCtx(ctx, &resp, query, tag)
	switch err {
	case nil:
		return &resp, nil
	case sqlx.ErrNotFound:
		return nil, ErrNotFound
	default:
		return nil, err
	}
}

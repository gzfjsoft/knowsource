package model

import (
	"context"
	"fmt"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ BlogCategoryModel = (*customBlogCategoryModel)(nil)

type (
	// BlogCategoryModel is an interface to be customized, add more methods here,
	// and implement the added methods in customBlogCategoryModel.
	BlogCategoryModel interface {
		blogCategoryModel
		withSession(session sqlx.Session) BlogCategoryModel
		FindAll(ctx context.Context) ([]*BlogCategory, error)
		FindOneByCategory(ctx context.Context, category string) (*BlogCategory, error)
	}

	customBlogCategoryModel struct {
		*defaultBlogCategoryModel
	}
)

// NewBlogCategoryModel returns a model for the database table.
func NewBlogCategoryModel(conn sqlx.SqlConn) BlogCategoryModel {
	return &customBlogCategoryModel{
		defaultBlogCategoryModel: newBlogCategoryModel(conn),
	}
}

func (m *customBlogCategoryModel) withSession(session sqlx.Session) BlogCategoryModel {
	return NewBlogCategoryModel(sqlx.NewSqlConnFromSession(session))
}

func (m *defaultBlogCategoryModel) FindAll(ctx context.Context) ([]*BlogCategory, error) {
	query := fmt.Sprintf("select %s from %s", blogCategoryRows, m.table)
	var resp []*BlogCategory
	err := m.conn.QueryRowsCtx(ctx, &resp, query)
	return resp, err
}

func (m *defaultBlogCategoryModel) FindOneByCategory(ctx context.Context, category string) (*BlogCategory, error) {
	query := fmt.Sprintf("select %s from %s where `category` = ? limit 1", blogCategoryRows, m.table)
	var resp BlogCategory
	err := m.conn.QueryRowCtx(ctx, &resp, query, category)
	switch err {
	case nil:
		return &resp, nil
	case sqlx.ErrNotFound:
		return nil, ErrNotFound
	default:
		return nil, err
	}
}

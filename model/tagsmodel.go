package model

import (
	"context"
	"fmt"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ TagsModel = (*customTagsModel)(nil)

type (
	TagsModel interface {
		tagsModel
		withSession(session sqlx.Session) TagsModel
		FindOneByClientIdTag(ctx context.Context, clientId, tag string) (*Tags, error)
		DeleteByClientIdTag(ctx context.Context, clientId, tag string) error
		UpdateTagByClientId(ctx context.Context, clientId, oldTag, newTag string) error
		FindAllWithCondition(ctx context.Context, clientId, tagLike string, offset, limit int64) ([]*Tags, error)
		CountWithCondition(ctx context.Context, clientId, tagLike string) (int64, error)
	}

	customTagsModel struct {
		*defaultTagsModel
	}
)

func NewTagsModel(conn sqlx.SqlConn) TagsModel {
	return &customTagsModel{
		defaultTagsModel: newTagsModel(conn),
	}
}

func (m *customTagsModel) FindOneByClientIdTag(ctx context.Context, clientId, tag string) (*Tags, error) {
	query := fmt.Sprintf("SELECT %s FROM %s WHERE `client_id` = ? AND `tag` = ? LIMIT 1", tagsRows, m.table)
	var resp Tags
	err := m.conn.QueryRowCtx(ctx, &resp, query, clientId, tag)
	switch err {
	case nil:
		return &resp, nil
	case sqlx.ErrNotFound:
		return nil, ErrNotFound
	default:
		return nil, err
	}
}

func (m *customTagsModel) DeleteByClientIdTag(ctx context.Context, clientId, tag string) error {
	query := fmt.Sprintf("DELETE FROM %s WHERE `client_id` = ? AND `tag` = ?", m.table)
	_, err := m.conn.ExecCtx(ctx, query, clientId, tag)
	return err
}

func (m *customTagsModel) UpdateTagByClientId(ctx context.Context, clientId, oldTag, newTag string) error {
	query := fmt.Sprintf("UPDATE %s SET `tag` = ? WHERE `client_id` = ? AND `tag` = ?", m.table)
	_, err := m.conn.ExecCtx(ctx, query, newTag, clientId, oldTag)
	return err
}

func (m *customTagsModel) withSession(session sqlx.Session) TagsModel {
	return NewTagsModel(sqlx.NewSqlConnFromSession(session))
}

func (m *customTagsModel) FindAllWithCondition(ctx context.Context, clientId, tag string, offset, limit int64) ([]*Tags, error) {
	query := fmt.Sprintf("SELECT %s FROM %s WHERE `client_id` = ?", tagsRows, m.table)
	args := []interface{}{clientId}
	if tag != "" {
		query += " AND `tag` LIKE ?"
		args = append(args, "%"+tag+"%")
	}
	query += " ORDER BY `tag` ASC LIMIT ? OFFSET ?"
	args = append(args, limit, offset)

	var tags []*Tags
	err := m.conn.QueryRowsCtx(ctx, &tags, query, args...)
	return tags, err
}

func (m *customTagsModel) CountWithCondition(ctx context.Context, clientId, tag string) (int64, error) {
	query := fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE `client_id` = ?", m.table)
	args := []interface{}{clientId}
	if tag != "" {
		query += " AND `tag` LIKE ?"
		args = append(args, "%"+tag+"%")
	}
	var count int64
	err := m.conn.QueryRowCtx(ctx, &count, query, args...)
	return count, err
}

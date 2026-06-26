package model

import (
	"context"
	"fmt"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ UserUploadPhotoModel = (*customUserUploadPhotoModel)(nil)

type (
	// UserUploadPhotoModel is an interface to be customized, add more methods here,
	// and implement the added methods in customUserUploadPhotoModel.
	UserUploadPhotoModel interface {
		userUploadPhotoModel
		withSession(session sqlx.Session) UserUploadPhotoModel
		FindByUserIdAndAuditStatus(ctx context.Context, userId uint64, auditStatus int64) ([]*UserUploadPhoto, error)
	}

	customUserUploadPhotoModel struct {
		*defaultUserUploadPhotoModel
	}
)

// NewUserUploadPhotoModel returns a model for the database table.
func NewUserUploadPhotoModel(conn sqlx.SqlConn) UserUploadPhotoModel {
	return &customUserUploadPhotoModel{
		defaultUserUploadPhotoModel: newUserUploadPhotoModel(conn),
	}
}

func (m *customUserUploadPhotoModel) withSession(session sqlx.Session) UserUploadPhotoModel {
	return NewUserUploadPhotoModel(sqlx.NewSqlConnFromSession(session))
}

// FindByUserIdAndAuditStatus returns all photos for a user with a specific audit status
func (m *customUserUploadPhotoModel) FindByUserIdAndAuditStatus(ctx context.Context, userId uint64, auditStatus int64) ([]*UserUploadPhoto, error) {
	query := fmt.Sprintf("select %s from %s where `user_id` = ? and `is_audited` = ? order by `created_at` desc", userUploadPhotoRows, m.table)
	var resp []*UserUploadPhoto
	err := m.conn.QueryRowsCtx(ctx, &resp, query, userId, auditStatus)
	return resp, err
}

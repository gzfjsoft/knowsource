package model

import "github.com/zeromicro/go-zero/core/stores/sqlx"

var _ MomentCommentLikeModel = (*customMomentCommentLikeModel)(nil)

type (
	// MomentCommentLikeModel is an interface to be customized, add more methods here,
	// and implement the added methods in customMomentCommentLikeModel.
	MomentCommentLikeModel interface {
		momentCommentLikeModel
		withSession(session sqlx.Session) MomentCommentLikeModel
	}

	customMomentCommentLikeModel struct {
		*defaultMomentCommentLikeModel
	}
)

// NewMomentCommentLikeModel returns a model for the database table.
func NewMomentCommentLikeModel(conn sqlx.SqlConn) MomentCommentLikeModel {
	return &customMomentCommentLikeModel{
		defaultMomentCommentLikeModel: newMomentCommentLikeModel(conn),
	}
}

func (m *customMomentCommentLikeModel) withSession(session sqlx.Session) MomentCommentLikeModel {
	return NewMomentCommentLikeModel(sqlx.NewSqlConnFromSession(session))
}

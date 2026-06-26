package model

import "github.com/zeromicro/go-zero/core/stores/sqlx"

var _ AiUsageStatsModel = (*customAiUsageStatsModel)(nil)

type (
	// AiUsageStatsModel is an interface to be customized, add more methods here,
	// and implement the added methods in customAiUsageStatsModel.
	AiUsageStatsModel interface {
		aiUsageStatsModel
		withSession(session sqlx.Session) AiUsageStatsModel
	}

	customAiUsageStatsModel struct {
		*defaultAiUsageStatsModel
	}
)

// NewAiUsageStatsModel returns a model for the database table.
func NewAiUsageStatsModel(conn sqlx.SqlConn) AiUsageStatsModel {
	return &customAiUsageStatsModel{
		defaultAiUsageStatsModel: newAiUsageStatsModel(conn),
	}
}

func (m *customAiUsageStatsModel) withSession(session sqlx.Session) AiUsageStatsModel {
	return NewAiUsageStatsModel(sqlx.NewSqlConnFromSession(session))
}

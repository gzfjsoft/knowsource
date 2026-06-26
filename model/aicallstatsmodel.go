package model

import (
	"context"
	"strings"
	"time"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ AiCallStatsModel = (*customAiCallStatsModel)(nil)

// AiCallStatsByTimeRange 按时间范围统计结果
type AiCallStatsByTimeRange struct {
	TotalCount           int64   `db:"total_count"`
	UserCount            int64   `db:"user_count"`
	ModelCount           int64   `db:"model_count"`
	AvgCostMs            float64 `db:"avg_cost_ms"`
	SumQuestionCharCount int64   `db:"sum_question_char_count"`
	SumOutputCharCount   int64   `db:"sum_output_char_count"`
	ModelNames           []string
}

type (
	// AiCallStatsModel is an interface to be customized, add more methods here,
	// and implement the added methods in customAiCallStatsModel.
	AiCallStatsModel interface {
		aiCallStatsModel
		withSession(session sqlx.Session) AiCallStatsModel
		StatsByTimeRange(ctx context.Context, clientId string, startTime, endTime time.Time) (*AiCallStatsByTimeRange, error)
	}

	customAiCallStatsModel struct {
		*defaultAiCallStatsModel
	}
)

// NewAiCallStatsModel returns a model for the database table.
func NewAiCallStatsModel(conn sqlx.SqlConn) AiCallStatsModel {
	return &customAiCallStatsModel{
		defaultAiCallStatsModel: newAiCallStatsModel(conn),
	}
}

func (m *customAiCallStatsModel) withSession(session sqlx.Session) AiCallStatsModel {
	return NewAiCallStatsModel(sqlx.NewSqlConnFromSession(session))
}

// statsByTimeRangeRow 仅用于第一次聚合查询的扫描目标（列数一致，避免 not matching destination to scan）
type statsByTimeRangeRow struct {
	TotalCount           int64   `db:"total_count"`
	UserCount            int64   `db:"user_count"`
	ModelCount           int64   `db:"model_count"`
	AvgCostMs            float64 `db:"avg_cost_ms"`
	SumQuestionCharCount int64   `db:"sum_question_char_count"`
	SumOutputCharCount   int64   `db:"sum_output_char_count"`
}

// StatsByTimeRange 按时间范围统计：总次数、去重用户数、模型种类数及模型名称列表
func (m *customAiCallStatsModel) StatsByTimeRange(ctx context.Context, clientId string, startTime, endTime time.Time) (*AiCallStatsByTimeRange, error) {
	res := &AiCallStatsByTimeRange{}
	// 统计总数、人数、模型数（endTime 使用闭区间，即当天结束前）
	query := `SELECT COUNT(*) AS total_count,
		COUNT(DISTINCT user_id) AS user_count,
		COUNT(DISTINCT model_name) AS model_count,
		COALESCE(AVG(cost_ms), 0) AS avg_cost_ms,
		COALESCE(SUM(question_char_count), 0) AS sum_question_char_count,
		COALESCE(SUM(output_char_count), 0) AS sum_output_char_count
		FROM ` + m.table + ` WHERE call_time >= ? AND call_time <= ?`
	args := []interface{}{startTime, endTime}
	if strings.TrimSpace(clientId) != "" {
		query += " AND client_id = ?"
		args = append(args, clientId)
	}
	var row statsByTimeRangeRow
	err := m.conn.QueryRowCtx(ctx, &row, query, args...)
	if err != nil {
		return nil, err
	}
	res.TotalCount = row.TotalCount
	res.UserCount = row.UserCount
	res.ModelCount = row.ModelCount
	res.AvgCostMs = row.AvgCostMs
	res.SumQuestionCharCount = row.SumQuestionCharCount
	res.SumOutputCharCount = row.SumOutputCharCount
	// 查询去重模型名称列表
	namesQuery := `SELECT DISTINCT model_name FROM ` + m.table + ` WHERE call_time >= ? AND call_time <= ?`
	nameArgs := []interface{}{startTime, endTime}
	if strings.TrimSpace(clientId) != "" {
		namesQuery += " AND client_id = ?"
		nameArgs = append(nameArgs, clientId)
	}
	namesQuery += " ORDER BY model_name"
	var nameRows []struct {
		ModelName string `db:"model_name"`
	}
	err = m.conn.QueryRowsCtx(ctx, &nameRows, namesQuery, nameArgs...)
	if err != nil {
		return res, nil // 已拿到聚合结果，名称列表失败时仅返回聚合
	}
	res.ModelNames = make([]string, 0, len(nameRows))
	for _, r := range nameRows {
		res.ModelNames = append(res.ModelNames, r.ModelName)
	}
	return res, nil
}

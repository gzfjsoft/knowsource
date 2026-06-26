// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package knowsource

import (
	"context"
	"strings"
	"time"

	"knowsource/api/internal/svc"
	"knowsource/api/internal/types"
	"knowsource/common/response"

	"github.com/zeromicro/go-zero/core/logx"
)

type AiCallStatsQueryLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// AI 调用统计：按时间范围查询，返回总次数、人数、模型数量
func NewAiCallStatsQueryLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AiCallStatsQueryLogic {
	return &AiCallStatsQueryLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *AiCallStatsQueryLogic) AiCallStatsQuery(req *types.AiCallStatsQueryRequest) (resp *types.AiCallStatsQueryResponse, err error) {
	clientId, _ := l.ctx.Value("clientId").(string)
	clientId = strings.TrimSpace(clientId)
	if clientId == "" {
		return &types.AiCallStatsQueryResponse{
			Response: types.Response{Code: response.UnauthorizedCode, Message: "clientId不能为空，请重新登录"},
		}, nil
	}

	// 解析时间，支持 2006-01-02 或 2006-01-02 15:04:05
	layouts := []string{"2006-01-02 15:04:05", "2006-01-02"}
	var startTime, endTime time.Time
	for _, layout := range layouts {
		if t, e := time.ParseInLocation(layout, req.StartTime, time.Local); e == nil {
			startTime = t
			break
		}
	}
	if startTime.IsZero() {
		return &types.AiCallStatsQueryResponse{
			Response: types.Response{Code: response.ParameterErrorCode, Message: "开始时间格式错误，请使用 2006-01-02 或 2006-01-02 15:04:05"},
		}, nil
	}
	for _, layout := range layouts {
		if t, e := time.ParseInLocation(layout, req.EndTime, time.Local); e == nil {
			endTime = t
			break
		}
	}
	if endTime.IsZero() {
		return &types.AiCallStatsQueryResponse{
			Response: types.Response{Code: response.ParameterErrorCode, Message: "结束时间格式错误，请使用 2006-01-02 或 2006-01-02 15:04:05"},
		}, nil
	}
	// 若结束时间仅为日期，则包含该日全天（到 23:59:59）
	if req.EndTime == endTime.Format("2006-01-02") {
		endTime = endTime.Add(24*time.Hour - time.Second)
	}
	if startTime.After(endTime) {
		return &types.AiCallStatsQueryResponse{
			Response: types.Response{Code: response.ParameterErrorCode, Message: "开始时间不能晚于结束时间"},
		}, nil
	}

	stats, err := l.svcCtx.AiCallStatsModel.StatsByTimeRange(l.ctx, clientId, startTime, endTime)
	if err != nil {
		logx.WithContext(l.ctx).Errorf("AiCallStatsModel.StatsByTimeRange: %v", err)
		return &types.AiCallStatsQueryResponse{
			Response: types.Response{Code: response.ServerErrorCode, Message: "Database error", Info: err.Error()},
		}, nil
	}
	return &types.AiCallStatsQueryResponse{
		Response: types.Response{Code: response.SuccessCode, Message: "success"},
		Data: &types.AiCallStatsQueryData{
			TotalCount:           stats.TotalCount,
			UserCount:            stats.UserCount,
			ModelCount:           stats.ModelCount,
			AvgCostMs:            stats.AvgCostMs,
			SumQuestionCharCount: stats.SumQuestionCharCount,
			SumOutputCharCount:   stats.SumOutputCharCount,
			ModelNames:           stats.ModelNames,
		},
	}, nil
}

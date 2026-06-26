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

type GetSystemStatsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 系统统计：返回员工总数、部门总数、知识库总数、AI 会话数
func NewGetSystemStatsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetSystemStatsLogic {
	return &GetSystemStatsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetSystemStatsLogic) GetSystemStats() (resp *types.SystemStatsResponse, err error) {
	var empCount, deptCount, documentCount, sessionCount int64
	//TODO: 暂时不加clientId过滤
	clientId, ok := l.ctx.Value("clientId").(string)
	if !ok {
		l.Errorf("获取clientId失败: %v", err)
		return &types.SystemStatsResponse{
			Response: types.Response{
				Code:    response.ServerErrorCode,
				Message: "GetSystemStats获取clientId失败",
				Info:    "获取clientId失败",
			},
		}, nil
	}
	clientId = strings.TrimSpace(clientId)
	if clientId == "" {
		// 兼容：若未注入 clientId，则返回全局统计（不加过滤）
		clientId = ""
	}

	// 查询员工总数
	if clientId == "" {
		if err := l.svcCtx.Mysql.QueryRowCtx(l.ctx, &empCount, "SELECT COUNT(*) FROM fr_emp"); err != nil {
			l.Errorf("查询员工总数失败: %v", err)
			empCount = 0
		}
	} else {
		if err := l.svcCtx.Mysql.QueryRowCtx(l.ctx, &empCount, "SELECT COUNT(*) FROM fr_emp WHERE client_id = ?", clientId); err != nil {
			l.Errorf("查询员工总数失败: %v", err)
			empCount = 0
		}
	}

	// 查询部门总数
	if clientId == "" {
		if err := l.svcCtx.Mysql.QueryRowCtx(l.ctx, &deptCount, "SELECT COUNT(*) FROM fr_dept"); err != nil {
			l.Errorf("查询部门总数失败: %v", err)
			deptCount = 0
		}
	} else {
		if err := l.svcCtx.Mysql.QueryRowCtx(l.ctx, &deptCount, "SELECT COUNT(*) FROM fr_dept WHERE client_id = ?", clientId); err != nil {
			l.Errorf("查询部门总数失败: %v", err)
			deptCount = 0
		}
	}

	// 查询知识库总数
	if clientId == "" {
		if err := l.svcCtx.Mysql.QueryRowCtx(l.ctx, &documentCount, "SELECT COUNT(*) FROM document_type"); err != nil {
			l.Errorf("查询知识库总数失败: %v", err)
			documentCount = 0
		}
	} else {
		if err := l.svcCtx.Mysql.QueryRowCtx(l.ctx, &documentCount, "SELECT COUNT(*) FROM document_type WHERE client_id = ?", clientId); err != nil {
			l.Errorf("查询知识库总数失败: %v", err)
			documentCount = 0
		}
	}

	// 查询AI会话数：使用 AiCallStatsModel.StatsByTimeRange 从 2026/1/1 到现在
	startTime := time.Date(2026, 1, 1, 0, 0, 0, 0, time.Local)
	endTime := time.Now()
	stats, err := l.svcCtx.AiCallStatsModel.StatsByTimeRange(l.ctx, clientId, startTime, endTime)
	if err != nil {
		l.Errorf("查询AI会话数失败: %v", err)
		sessionCount = 0
	} else {
		sessionCount = stats.TotalCount
	}

	return &types.SystemStatsResponse{
		Response: types.Response{
			Code:    200,
			Message: "success",
		},
		Data: &types.SystemStatsData{
			EmpCount:      empCount,
			DeptCount:     deptCount,
			DocumentCount: documentCount,
			SessionCount:  sessionCount,
		},
	}, nil
}

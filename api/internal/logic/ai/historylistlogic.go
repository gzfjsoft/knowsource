package ai

import (
	"context"

	"knowsource/api/internal/svc"
	"knowsource/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type HistoryListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewHistoryListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *HistoryListLogic {
	return &HistoryListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *HistoryListLogic) HistoryList() (resp *types.HistoryListResponse, err error) {
	// 从上下文获取用户ID，这里需要根据实际的认证方式获取
	// 假设从JWT token或其他方式获取用户ID

	empcode, _ := l.ctx.Value("empCode").(string)
	clientId, _ := l.ctx.Value("clientId").(string)
	if empcode == "" || clientId == "" {
		return &types.HistoryListResponse{
			Response: types.Response{
				Code:    400,
				Message: "登录信息不完整",
			},
		}, nil
	}
	// 查询用户的会话列表，包含最后一条消息
	sessions, err := l.svcCtx.AiSessionsModel.FindByEmpCodeWithLastMessage(l.ctx, clientId, empcode)
	if err != nil {
		l.Errorf("查询会话列表失败: %v", err)
		return &types.HistoryListResponse{
			Response: types.Response{
				Code:    500,
				Message: "查询会话列表失败",
			},
		}, nil
	}

	// 转换为API响应格式
	sessionInfos := make([]types.SessionInfo, 0, len(sessions))
	for _, session := range sessions {
		var updateTime int64
		if session.LastMessageTime.Valid {
			updateTime = session.LastMessageTime.Time.Unix()
		} else {
			updateTime = session.CreatedAt.Unix()
		}

		sessionInfos = append(sessionInfos, types.SessionInfo{
			Session:          session.SessionUuid,
			CategoryId:       session.CategoryId,
			EmpCode:          session.EmpCode,
			Keys:             session.Keys,
			Title:            session.Title,
			DocumentTypeCode: session.DocumentTypeCode,
			LastQuery:        session.LastQuery,
			LastReply:        session.LastReply,
			Tags:             session.Tags,
			UpdateTime:       updateTime,
		})
	}

	return &types.HistoryListResponse{
		Response: types.Response{
			Code:    200,
			Message: "success",
		},
		Sessions: sessionInfos,
	}, nil
}

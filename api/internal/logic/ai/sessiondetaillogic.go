package ai

import (
	"context"
	"encoding/json"
	"strings"

	"knowsource/api/internal/svc"
	"knowsource/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type SessionDetailLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewSessionDetailLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SessionDetailLogic {
	return &SessionDetailLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *SessionDetailLogic) SessionDetail(req *types.SessionDetailRequest) (resp *types.SessionDetailResponse, err error) {
	// 验证请求参数
	if req.Session == "" {
		return &types.SessionDetailResponse{
			Response: types.Response{
				Code:    400,
				Message: "会话ID不能为空",
			},
		}, nil
	}

	// 从上下文获取用户ID
	empcode := l.ctx.Value("empCode").(string)
	if empcode == "" {
		return &types.SessionDetailResponse{
			Response: types.Response{
				Code:    400,
				Message: "员工编码不能为空",
			},
		}, nil
	}

	clientId, _ := l.ctx.Value("clientId").(string)
	if strings.TrimSpace(clientId) == "" {
		return &types.SessionDetailResponse{
			Response: types.Response{
				Code:    401,
				Message: "clientId不能为空，请重新登录",
			},
		}, nil
	}

	// 首先验证会话是否存在且属于当前用户
	session, err := l.svcCtx.AiSessionsModel.FindOneByClientIdSessionUuid(l.ctx, clientId, req.Session)
	if err != nil {
		l.Errorf("查询会话失败: %v", err)
		return &types.SessionDetailResponse{
			Response: types.Response{
				Code:    500,
				Message: "查询会话失败",
			},
		}, nil
	}

	// 检查会话是否属于当前用户
	if session.EmpCode != empcode {
		return &types.SessionDetailResponse{
			Response: types.Response{
				Code:    403,
				Message: "无权限访问他人会话",
			},
		}, nil
	}

	// 查询会话的消息列表
	messages, err := l.svcCtx.AiMessagesModel.FindBySessionUuid(l.ctx, req.Session)
	if err != nil {
		l.Errorf("查询消息列表失败: %v", err)
		return &types.SessionDetailResponse{
			Response: types.Response{
				Code:    500,
				Message: "查询消息列表失败",
			},
		}, nil
	}

	// 转换为API响应格式
	aiMessages := make([]types.AIMessage, 0, len(messages))
	for _, msg := range messages {
		var createdAt int64
		if msg.CreatedAtUnix > 0 {
			createdAt = msg.CreatedAtUnix
		} else {
			createdAt = msg.CreatedAt.Unix()
		}

		aiMessage := types.AIMessage{
			Role:      msg.Role,
			Content:   msg.Content,
			CreatedAt: createdAt,
		}

		// 添加思考过程（如果存在）
		if msg.Thinking.Valid && msg.Thinking.String != "" {
			aiMessage.Thinking = msg.Thinking.String
		}

		// 解析并填充 files（参考资料/附件）
		if msg.Files != "" {
			var files []string
			if err := json.Unmarshal([]byte(msg.Files), &files); err == nil && len(files) > 0 {
				aiMessage.Files = files
			}
		}

		aiMessages = append(aiMessages, aiMessage)
	}

	return &types.SessionDetailResponse{
		Response: types.Response{
			Code:    200,
			Message: "success",
		},
		Messages: aiMessages,
	}, nil
}

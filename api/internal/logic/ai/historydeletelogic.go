package ai

import (
	"context"
	"errors"
	"strings"

	"knowsource/api/internal/svc"
	"knowsource/api/internal/types"
	"knowsource/model"

	"github.com/zeromicro/go-zero/core/logx"
)

type HistoryDeleteLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 删除历史会话
func NewHistoryDeleteLogic(ctx context.Context, svcCtx *svc.ServiceContext) *HistoryDeleteLogic {
	return &HistoryDeleteLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *HistoryDeleteLogic) HistoryDelete(req *types.SessionDeleteRequest) (resp *types.Response, err error) {
	// 验证请求参数
	if req.Session == "" {
		return &types.Response{
			Code:    400,
			Message: "会话ID不能为空",
		}, nil
	}

	// 从上下文获取用户ID
	empcode, ok := l.ctx.Value("empCode").(string)
	if !ok || empcode == "" {
		return &types.Response{
			Code:    400,
			Message: "员工编码不能为空",
		}, nil
	}

	clientId, _ := l.ctx.Value("clientId").(string)
	if strings.TrimSpace(clientId) == "" {
		return &types.Response{
			Code:    401,
			Message: "clientId不能为空，请重新登录",
		}, nil
	}

	// 查找会话
	session, err := l.svcCtx.AiSessionsModel.FindOneByClientIdSessionUuid(l.ctx, clientId, req.Session)
	if err != nil {
		if errors.Is(err, model.ErrNotFound) {
			return &types.Response{
				Code:    404,
				Message: "会话不存在",
			}, nil
		}
		l.Errorf("查询会话失败: %v", err)
		return &types.Response{
			Code:    500,
			Message: "查询会话失败",
			Info:    err.Error(),
		}, nil
	}

	// 检查会话是否属于当前用户
	if session.EmpCode != empcode {
		return &types.Response{
			Code:    403,
			Message: "无权限删除他人会话",
		}, nil
	}

	// 删除会话（由于外键约束设置了CASCADE，相关的消息会自动删除）
	err = l.svcCtx.AiSessionsModel.Delete(l.ctx, session.SessionId)
	if err != nil {
		l.Errorf("删除会话失败: %v", err)
		return &types.Response{
			Code:    500,
			Message: "删除会话失败",
			Info:    err.Error(),
		}, nil
	}

	l.Infof("成功删除会话: session_uuid=%s, emp_code=%s", req.Session, empcode)

	return &types.Response{
		Code:    200,
		Message: "删除成功",
	}, nil
}

package ai

import (
	"context"
	"os"
	"path/filepath"
	"strings"

	"knowsource/api/internal/svc"
	"knowsource/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetAiLogLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 管理员获取ailog.txt
func NewGetAiLogLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetAiLogLogic {
	return &GetAiLogLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetAiLogLogic) GetAiLog(req *types.PathNameRequest) (resp *types.GetAiLogResponse, err error) {
	// 从 ai_session_logs 目录及其子目录（按 clientid 分类）获取会话日志文件
	logDir := "ai_session_logs"

	// 获取当前用户的 clientid
	clientId, _ := l.ctx.Value("clientId").(string)
	if clientId == "" {
		return &types.GetAiLogResponse{
			Response: types.Response{
				Code:    401,
				Message: "未获取到 clientId，请重新登录",
			},
		}, nil
	}

	// 若 name 已包含 .log.txt 则不再拼接
	clientLogDir := filepath.Join(logDir, clientId)
	latestFile := filepath.Join(clientLogDir, req.Name+".log.txt")
	latestFile = strings.ReplaceAll(latestFile, ".log.txt.log.txt", ".log.txt")

	// 读取文件内容
	content, err := os.ReadFile(latestFile)
	if err != nil {
		l.Logger.Errorf("读取日志文件失败: %v", err)
		return &types.GetAiLogResponse{
			Response: types.Response{
				Code:    500,
				Message: "读取日志文件失败: " + err.Error(),
			},
		}, nil
	}

	return &types.GetAiLogResponse{
		Response: types.Response{
			Code:    200,
			Message: "success",
		},
		Data: &types.GetAiLogResponseData{
			Content: string(content),
		},
	}, nil
}

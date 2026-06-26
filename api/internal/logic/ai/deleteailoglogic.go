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

type DeleteAiLogLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 管理员删除ailog.txt
func NewDeleteAiLogLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteAiLogLogic {
	return &DeleteAiLogLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *DeleteAiLogLogic) DeleteAiLog(req *types.PathNameRequest) (resp *types.Response, err error) {
	// 从 ai_session_logs 目录及其子目录（按 clientid 分类）删除日志文件
	logDir := "ai_session_logs"

	// 获取当前用户的 clientid
	clientId, _ := l.ctx.Value("clientId").(string)
	if clientId == "" {
		return &types.Response{
			Code:    401,
			Message: "未获取到 clientId，请重新登录",
		}, nil
	}

	clientLogDir := filepath.Join(logDir, clientId)
	filePath := filepath.Join(clientLogDir, req.Name+".log.txt")
	filePath = strings.ReplaceAll(filePath, ".log.txt.log.txt", ".log.txt")

	// 检查文件是否存在
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		l.Logger.Errorf("日志文件不存在: %s", filePath)
		return &types.Response{
			Code:    404,
			Message: "日志文件不存在",
		}, nil
	}

	// 删除文件
	err = os.Remove(filePath)
	if err != nil {
		l.Logger.Errorf("删除日志文件失败: %v", err)
		return &types.Response{
			Code:    500,
			Message: "删除日志文件失败: " + err.Error(),
		}, nil
	}

	l.Logger.Infof("成功删除日志文件: %s", filePath)
	return &types.Response{
		Code:    200,
		Message: "success",
	}, nil
}

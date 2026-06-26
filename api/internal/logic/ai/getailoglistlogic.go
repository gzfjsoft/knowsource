package ai

import (
	"context"
	"os"
	"path/filepath"
	"sort"
	"time"

	"knowsource/api/internal/svc"
	"knowsource/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetAiLogListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 管理员获取ailog.txt list
func NewGetAiLogListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetAiLogListLogic {
	return &GetAiLogListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetAiLogListLogic) GetAiLogList() (resp *types.GetAiLogListResponse, err error) {
	// 从 ai_session_logs 目录及其子目录（按 clientid 分类）查找所有 *.log.txt 文件
	logDir := "ai_session_logs"

	// 获取当前用户的 clientid
	clientId, _ := l.ctx.Value("clientId").(string)
	if clientId == "" {
		return &types.GetAiLogListResponse{
			Response: types.Response{
				Code:    401,
				Message: "未获取到 clientId，请重新登录",
			},
		}, nil
	}

	// 查找当前 clientid 目录下的所有日志文件
	clientLogDir := filepath.Join(logDir, clientId)
	pattern := filepath.Join(clientLogDir, "*.log.txt")
	matches, err := filepath.Glob(pattern)
	if err != nil {
		l.Logger.Errorf("查找日志文件失败: %v", err)
		return &types.GetAiLogListResponse{
			Response: types.Response{
				Code:    500,
				Message: "查找日志文件失败: " + err.Error(),
			},
		}, nil
	}

	// 创建文件信息结构体，包含文件名和修改时间
	type fileInfo struct {
		name    string
		modTime time.Time
	}

	fileInfos := make([]fileInfo, 0, len(matches))
	for _, match := range matches {
		// 获取文件信息
		info, err := os.Stat(match)
		if err != nil {
			l.Logger.Errorf("获取文件信息失败: %v, 文件: %s", err, match)
			continue
		}

		fileInfos = append(fileInfos, fileInfo{
			name:    filepath.Base(match),
			modTime: info.ModTime(),
		})
	}

	// 按修改时间降序排序（最新的在前）
	sort.Slice(fileInfos, func(i, j int) bool {
		return fileInfos[i].modTime.After(fileInfos[j].modTime)
	})

	list := make([]types.AiLogFileItem, 0, len(fileInfos))
	for _, info := range fileInfos {
		list = append(list, types.AiLogFileItem{
			Name:     info.name,
			DateTime: info.modTime.Unix(),
		})
	}

	return &types.GetAiLogListResponse{
		Response: types.Response{
			Code:    200,
			Message: "success",
		},
		Data: &types.GetAiLogListResponseData{
			List: list,
		},
	}, nil
}

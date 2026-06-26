// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package ai

import (
	"context"
	"os"
	"path/filepath"
	"sort"
	"time"

	"knowsource/api/internal/svc"
	"knowsource/api/internal/types"
	"knowsource/api/internal/utils"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetRawDocQaLlmLogListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 管理员获取问答抽取LLM日志列表
func NewGetRawDocQaLlmLogListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetRawDocQaLlmLogListLogic {
	return &GetRawDocQaLlmLogListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetRawDocQaLlmLogListLogic) GetRawDocQaLlmLogList() (resp *types.GetRawDocQaLlmLogListResponse, err error) {
	clientId, _ := l.ctx.Value("clientId").(string)
	if clientId == "" {
		return &types.GetRawDocQaLlmLogListResponse{
			Response: types.Response{
				Code:    401,
				Message: "未获取到 clientId，请重新登录",
			},
		}, nil
	}

	clientLogDir := filepath.Join(utils.RawDocQaLLMLogDir, clientId)
	pattern := filepath.Join(clientLogDir, "*.log.txt")
	matches, globErr := filepath.Glob(pattern)
	if globErr != nil {
		return &types.GetRawDocQaLlmLogListResponse{
			Response: types.Response{
				Code:    500,
				Message: "查找日志文件失败: " + globErr.Error(),
			},
		}, nil
	}

	type fileInfo struct {
		name    string
		modTime time.Time
	}
	items := make([]fileInfo, 0, len(matches))
	for _, match := range matches {
		info, statErr := os.Stat(match)
		if statErr != nil {
			continue
		}
		items = append(items, fileInfo{name: filepath.Base(match), modTime: info.ModTime()})
	}
	sort.Slice(items, func(i, j int) bool {
		return items[i].modTime.After(items[j].modTime)
	})
	list := make([]types.AiLogFileItem, 0, len(items))
	for _, item := range items {
		list = append(list, types.AiLogFileItem{
			Name:     item.name,
			DateTime: item.modTime.Unix(),
		})
	}

	return &types.GetRawDocQaLlmLogListResponse{
		Response: types.Response{
			Code:    200,
			Message: "success",
		},
		Data: &types.GetRawDocQaLlmLogListResponseData{
			List: list,
		},
	}, nil
}

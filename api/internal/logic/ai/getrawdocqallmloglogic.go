// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package ai

import (
	"context"
	"os"
	"path/filepath"
	"strings"

	"knowsource/api/internal/svc"
	"knowsource/api/internal/types"
	"knowsource/api/internal/utils"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetRawDocQaLlmLogLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 管理员获取问答抽取LLM日志内容
func NewGetRawDocQaLlmLogLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetRawDocQaLlmLogLogic {
	return &GetRawDocQaLlmLogLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetRawDocQaLlmLogLogic) GetRawDocQaLlmLog(req *types.PathNameRequest) (resp *types.GetAiLogResponse, err error) {
	clientId, _ := l.ctx.Value("clientId").(string)
	if clientId == "" {
		return &types.GetAiLogResponse{
			Response: types.Response{
				Code:    401,
				Message: "未获取到 clientId，请重新登录",
			},
		}, nil
	}
	name := strings.TrimSpace(req.Name)
	if name == "" {
		return &types.GetAiLogResponse{
			Response: types.Response{
				Code:    400,
				Message: "name 不能为空",
			},
		}, nil
	}
	name = filepath.Base(name)
	if !strings.HasSuffix(name, ".log.txt") {
		name += ".log.txt"
	}
	logFile := filepath.Join(utils.RawDocQaLLMLogDir, clientId, name)
	content, readErr := os.ReadFile(logFile)
	if readErr != nil {
		return &types.GetAiLogResponse{
			Response: types.Response{
				Code:    500,
				Message: "读取日志文件失败: " + readErr.Error(),
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

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

type DeleteRawDocQaLlmLogLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 管理员删除问答抽取LLM日志
func NewDeleteRawDocQaLlmLogLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteRawDocQaLlmLogLogic {
	return &DeleteRawDocQaLlmLogLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *DeleteRawDocQaLlmLogLogic) DeleteRawDocQaLlmLog(req *types.PathNameRequest) (resp *types.Response, err error) {
	clientId, _ := l.ctx.Value("clientId").(string)
	if clientId == "" {
		return &types.Response{
			Code:    401,
			Message: "未获取到 clientId，请重新登录",
		}, nil
	}
	name := strings.TrimSpace(req.Name)
	if name == "" {
		return &types.Response{
			Code:    400,
			Message: "name 不能为空",
		}, nil
	}
	name = filepath.Base(name)
	if !strings.HasSuffix(name, ".log.txt") {
		name += ".log.txt"
	}
	logFile := filepath.Join(utils.RawDocQaLLMLogDir, clientId, name)
	if _, statErr := os.Stat(logFile); os.IsNotExist(statErr) {
		return &types.Response{
			Code:    404,
			Message: "日志文件不存在",
		}, nil
	}
	if delErr := os.Remove(logFile); delErr != nil {
		return &types.Response{
			Code:    500,
			Message: "删除日志文件失败: " + delErr.Error(),
		}, nil
	}
	return &types.Response{
		Code:    200,
		Message: "success",
	}, nil
}

package knowdata

import (
	"context"
	"knowsource/common/response"
	"knowsource/model"

	"github.com/pkg/errors"

	"knowsource/api/internal/svc"
	"knowsource/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type DeleteAIConfigLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewDeleteAIConfigLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteAIConfigLogic {
	return &DeleteAIConfigLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *DeleteAIConfigLogic) DeleteAIConfig(req *types.PathIdRequest) (resp *response.Response, err error) {
	item, err := l.svcCtx.AiConfigModel.FindOne(l.ctx, req.Id)
	if err != nil {
		if errors.Is(err, model.ErrNotFound) {
			// 记录不存在的情况
			l.Logger.Info("AI配置记录不存在", logx.Field("id", req.Id))
			return &response.Response{
				Code:    response.RecordNotExistCode,
				Message: "记录不存在或已删除",
			}, nil
		}
		// 其他查询错误
		l.Logger.Error("查询AI配置失败", logx.Field("id", req.Id), logx.Field("error", err))
		return &response.Response{
			Code:    response.ServerErrorCode,
			Message: err.Error(),
		}, nil
	}

	// 	rag.js：调用 RAG 逻辑脚本
	// 检索提示词：拿到 RAG 数据后拼接发给 LLM 的信息格式
	// prompt：给 LLM 的提示词
	// greet：进入打招呼的语句
	// model：LLM 使用的模型

	var targetNames = map[string]bool{

		"检索提示词":   true,
		"角色提示词":   true,
		"问候词":     true,
		"问答提取提示词": true,
	}

	if item.DocumentCode == "" {
		if targetNames[item.Name] {
			return &response.Response{
				Code:    response.ConflictCode,
				Message: "基础配置不能删除",
			}, nil
		}
	}

	err = l.svcCtx.AiConfigModel.Delete(l.ctx, req.Id)
	if err != nil {
		l.Logger.Errorf("删除AI配置失败，ID: %d, 错误: %v", req.Id, err)
		return &response.Response{
			Code:    response.ServerErrorCode,
			Message: err.Error(),
		}, nil
	}

	// 删除成功日志
	userName, ok := l.ctx.Value("userName").(string)
	if !ok || userName == "" {
		userName = "unknown" // 默认用户名
	}
	l.Logger.Info("删除AI配置成功", logx.Field("id", req.Id), logx.Field("userName", userName))

	return &response.Response{
		Code:    response.SuccessCode,
		Message: "success",
	}, nil
}

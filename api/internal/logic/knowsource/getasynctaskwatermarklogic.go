// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package knowsource

import (
	"context"
	"strings"

	"knowsource/api/internal/svc"
	"knowsource/api/internal/types"
	"knowsource/common/asynctasksignal"
	"knowsource/common/response"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/redis"
)

type GetAsyncTaskWatermarkLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// async_task 变更水印（按租户 Redis），前端轮询与本地比较，有变化再拉列表
func NewGetAsyncTaskWatermarkLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetAsyncTaskWatermarkLogic {
	return &GetAsyncTaskWatermarkLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetAsyncTaskWatermarkLogic) GetAsyncTaskWatermark(req *types.GetAsyncTaskWatermarkRequest) (resp *types.GetAsyncTaskWatermarkResponse, err error) {
	clientId, _ := l.ctx.Value("clientId").(string)
	clientId = strings.TrimSpace(clientId)
	if clientId == "" {
		return &types.GetAsyncTaskWatermarkResponse{
			Response: types.Response{Code: response.UnauthorizedCode, Message: "clientId不能为空，请重新登录"},
		}, nil
	}
	if l.svcCtx.RedisClient == nil {
		return &types.GetAsyncTaskWatermarkResponse{
			Response: types.Response{Code: response.ServerErrorCode, Message: "Redis 未配置"},
		}, nil
	}
	key := asynctasksignal.ClientWatermarkKey(clientId)
	val, e := l.svcCtx.RedisClient.GetCtx(l.ctx, key)
	if e != nil && e != redis.Nil {
		return &types.GetAsyncTaskWatermarkResponse{
			Response: types.Response{Code: response.ServerErrorCode, Message: "读取水印失败", Info: e.Error()},
		}, nil
	}
	if e == redis.Nil {
		val = ""
	}
	return &types.GetAsyncTaskWatermarkResponse{
		Response: types.Response{Code: response.SuccessCode, Message: "success"},
		Data: &types.GetAsyncTaskWatermarkData{
			Watermark: strings.TrimSpace(val),
		},
	}, nil
}

package ai

import (
	"context"

	"knowsource/api/internal/svc"
	"knowsource/api/internal/types"
	"knowsource/api/internal/utils"
	"knowsource/common/response"
	"knowsource/consts"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetQdrantCollectionListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 超级管理员获取Qdrant集合列表
func NewGetQdrantCollectionListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetQdrantCollectionListLogic {
	return &GetQdrantCollectionListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetQdrantCollectionListLogic) GetQdrantCollectionList() (resp *types.GetQdrantCollectionListResponse, err error) {
	newResp := func(code int64, message, info string) *types.GetQdrantCollectionListResponse {
		return &types.GetQdrantCollectionListResponse{
			Response: types.Response{
				Code:    code,
				Message: message,
				Info:    info,
			},
		}
	}

	// 检查是否为超级管理员
	role, _ := l.ctx.Value("role").(string)
	if role != consts.SUPER_ADMIN {
		return newResp(response.UnauthorizedCode, "没有权限", "只有超级管理员才能访问此接口"), nil
	}

	cfg := l.svcCtx.Config.Qdrant
	if cfg.Host == "" || cfg.Port == 0 {
		return newResp(response.ServerErrorCode, "Qdrant 配置缺失", "请在配置文件中配置 Qdrant.host 和 Qdrant.port"), nil
	}

	qc, qErr := utils.NewQdrantToolsFromConfig(&l.svcCtx.Config)
	if qErr != nil {
		l.Errorf("初始化 QdrantTools 失败: %v", qErr)
		return newResp(response.ServerErrorCode, "初始化 QdrantTools 失败", qErr.Error()), nil
	}

	names, err := qc.ListCollections(l.ctx)
	if err != nil {
		l.Errorf("获取 Qdrant collection 列表失败: %v", err)
		return newResp(response.ServerErrorCode, "获取 Qdrant collection 列表失败", err.Error()), nil
	}

	return &types.GetQdrantCollectionListResponse{
		Response: types.Response{
			Code:    response.SuccessCode,
			Message: "success",
		},
		Data: &types.GetQdrantCollectionListResponseData{
			List: names,
		},
	}, nil
}

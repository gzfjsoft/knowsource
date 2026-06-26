package knowsource

import (
	"context"

	"knowsource/api/internal/svc"
	"knowsource/api/internal/types"
	"knowsource/api/internal/utils"
	"knowsource/common/response"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetQdrantCollectionListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 获取 qrdant collection list
func NewGetQdrantCollectionListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetQdrantCollectionListLogic {
	return &GetQdrantCollectionListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetQdrantCollectionListLogic) GetQdrantCollectionList(req *types.QdrantCollectionListRequest) (resp *types.QdrantCollectionListResponse, err error) {
	newResp := func(code int64, message, info string) *types.QdrantCollectionListResponse {
		return &types.QdrantCollectionListResponse{
			Response: types.Response{
				Code:    code,
				Message: message,
				Info:    info,
			},
		}
	}

	cfg := l.svcCtx.Config.Qdrant
	if cfg.Host == "" || cfg.Port == 0 {
		return newResp(response.ServerErrorCode, "Qdrant 配置缺失", "请在 knowsource.yaml 中配置 Qdrant.host 和 Qdrant.port"), nil
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

	return &types.QdrantCollectionListResponse{
		Response: types.Response{
			Code:    response.SuccessCode,
			Message: "success",
		},
		Data: &types.QdrantCollectionListData{
			List:  names,
			Total: int64(len(names)),
		},
	}, nil
}

package knowsource

import (
	"context"
	"strings"

	"knowsource/api/internal/svc"
	"knowsource/api/internal/types"
	"knowsource/common/response"
	"knowsource/model"

	"github.com/pkg/errors"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type GetFrTagLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 获取标签
func NewGetFrTagLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetFrTagLogic {
	return &GetFrTagLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetFrTagLogic) GetFrTag(req *types.FrTagGetRequest) (resp *types.FrTagGetResponse, err error) {
	clientId, _ := l.ctx.Value("clientId").(string)
	clientId = strings.TrimSpace(clientId)
	if clientId == "" {
		return &types.FrTagGetResponse{
			Response: types.Response{
				Code:    response.UnauthorizedCode,
				Message: "clientId不能为空，请重新登录",
			},
		}, nil
	}

	if req.Tag == "" {
		return &types.FrTagGetResponse{
			Response: types.Response{
				Code:    response.ParameterErrorCode,
				Message: "标签名称不能为空",
			},
		}, nil
	}

	tag, err := l.svcCtx.FrTagsModel.FindOneByClientIdTag(l.ctx, clientId, req.Tag)
	if err != nil {
		if errors.Is(err, model.ErrNotFound) || err == sqlx.ErrNotFound {
			return &types.FrTagGetResponse{
				Response: types.Response{
					Code:    response.RecordNotExistCode,
					Message: "标签不存在",
				},
			}, nil
		}
		l.Logger.Errorf("查询标签失败: %v, Tag: %s", err, req.Tag)
		return &types.FrTagGetResponse{
			Response: types.Response{
				Code:    response.ServerErrorCode,
				Message: "Database error",
				Info:    err.Error(),
			},
		}, nil
	}

	return &types.FrTagGetResponse{
		Response: types.Response{
			Code:    response.SuccessCode,
			Message: "success",
		},
		Data: &types.FrTagInfo{
			Tag: tag.Tag,
		},
	}, nil
}

package logic

import (
	"context"
	"encoding/json"

	"knowsource/api/internal/svc"
	"knowsource/api/internal/types"
	"knowsource/common/response"
	"knowsource/model"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetVerifyIdcardLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetVerifyIdcardLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetVerifyIdcardLogic {
	return &GetVerifyIdcardLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetVerifyIdcardLogic) GetVerifyIdcard(req *types.GetVerifyIdcardRequest) (resp *types.GetVerifyIdcardResponse, err error) {
	uid, _ := l.ctx.Value("uid").(json.Number).Int64()

	userAuthIdcard, err := l.svcCtx.UserAuthIdcardModel.GetByUserId(l.ctx, (uid))
	if err != nil {
		if err == model.ErrNotFound {
			return &types.GetVerifyIdcardResponse{
				Response: types.Response{Message: "user auth idcard not found", Code: response.SuccessCode},
			}, nil
		}
		return &types.GetVerifyIdcardResponse{
			Response: types.Response{
				Message: "get user auth idcard failed",
				Code:    response.ServerErrorCode,
				Info:    err.Error(),
			},
		}, nil
	}

	return &types.GetVerifyIdcardResponse{
		Response: types.Response{Message: "user auth idcard found", Code: response.SuccessCode},
		Data: &types.GetVerifyIdcardData{
			Id:          userAuthIdcard.Id,
			IdcardFront: userAuthIdcard.ImageFront,
			IdcardBack:  userAuthIdcard.ImageBack,
			Name:        userAuthIdcard.ReqName,
			Idcard:      userAuthIdcard.ReqIdcardNum,
		},
	}, nil
}

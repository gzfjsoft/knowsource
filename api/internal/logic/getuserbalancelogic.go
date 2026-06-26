package logic

import (
	"context"
	"encoding/json"
	"fmt"
	"knowsource/api/internal/types"
	"knowsource/common/response"

	"knowsource/api/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetUserBalanceLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetUserBalanceLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetUserBalanceLogic {
	return &GetUserBalanceLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetUserBalanceLogic) GetUserBalance() (resp *types.UserBalanceResponse) {
	resp = &types.UserBalanceResponse{
		Response: types.Response{
			Code:    response.SuccessCode,
			Message: "Success",
		},
	}

	uid, _ := l.ctx.Value("uid").(json.Number).Int64()
	condition := fmt.Sprintf("where user_id = %d", uid)
	items, err := l.svcCtx.BalancesModel.FindListView(l.ctx, condition)
	if err != nil {
		resp.Response.Code = response.ServerErrorCode
		resp.Response.Message = err.Error()
		return resp
	}

	list := *items
	data := make([]types.UserBalance, len(list))
	for i := 0; i < len(list); i++ {
		data[i] = types.UserBalance{
			OrgId:        list[i].OrgId,
			OrgName:      list[i].OrgName,
			Balance:      list[i].Balance,
			CurrencyCode: list[i].CurrencyCode,
		}
	}
	resp.Data = data
	return resp
}

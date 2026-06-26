package logic

import (
	"context"
	"encoding/json"
	"knowsource/common/response"

	"knowsource/api/internal/svc"
	"knowsource/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListTransactionRecordsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewListTransactionRecordsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListTransactionRecordsLogic {
	return &ListTransactionRecordsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ListTransactionRecordsLogic) ListTransactionRecords(req *types.TransactionRecordsListRequest) response.Response {
	// get uid from context
	// uid := l.ctx.Value("uid").(uint64)
	uid, _ := l.ctx.Value("uid").(json.Number).Int64()
	if uid <= 0 {
		return response.Fail(response.UnauthorizedCode, "Unauthorized")
	}

	values := make([]interface{}, 0)
	condition := "where 1=1"
	if uid > 0 {
		condition += " and user_id = ?"
		values = append(values, uid)
	}
	if req.OrgId > 0 {
		condition += " and org_id = ?"
		values = append(values, req.OrgId)
	}
	if req.TransType > 0 {
		condition += " and trans_type = ?"
		values = append(values, req.TransType)
	}
	if req.CurrencyCode != "" {
		condition += " and currency_code = ?"
		values = append(values, req.CurrencyCode)
	}
	if req.OrderNo != "" {
		condition += " and order_no = ?"
		values = append(values, req.OrderNo)
	}
	if req.BeginCreatedAt > 0 {
		condition += " and created_at > ?"
		values = append(values, req.BeginCreatedAt)
	}
	if req.EndCreatedAt > 0 {
		condition += " and created_at < ?"
		values = append(values, req.EndCreatedAt)
	}

	total, err := l.svcCtx.TransactionRecordsModel.Count(l.ctx, condition, values...)
	if err != nil {
		return response.Fail(response.ServerErrorCode, err.Error())
	}
	condition += " order by created_at desc limit ?,?"
	values = append(values, (req.Page-1)*req.PageSize)
	values = append(values, req.PageSize)

	data, err := l.svcCtx.TransactionRecordsModel.FindList(l.ctx, condition, values...)
	if err != nil {
		return response.Fail(response.ServerErrorCode, err.Error())
	}

	// bytes, err := json.Marshal(data)
	// if err != nil {
	// 	return response.Fail(response.ServerErrorCode, err.Error())
	// }

	// var items []types.TransactionRecords
	// err = json.Unmarshal(bytes, &items)
	// if err != nil {
	// 	return response.Fail(response.ServerErrorCode, err.Error())
	// }

	// Id           uint64 `json:"id"`
	// UserId       uint64 `json:"userId"`
	// OrgId        uint64 `json:"orgId"`
	// CreatedAt    uint64 `json:"createdAt"`
	// TransType    uint64 `json:"transType"`
	// PayType      string `json:"payType"`
	// CurrencyCode string `json:"currencyCode"`
	// Detail       string `json:"detail"`
	// OrderNo      string `json:"orderNo"`
	// Username     string `json:"username"`
	// Amount       int64  `json:"amount"`
	// Balance      int64  `json:"balance"`

	var items []types.TransactionRecords
	for _, item := range *data {
		items = append(items, types.TransactionRecords{
			OrderNo:      item.OrderNo.String,
			CreatedAt:    uint64(item.CreatedAt.Unix()),
			Amount:       item.Amount,
			Balance:      item.Balance,
			TransType:    item.TransType,
			CurrencyCode: item.CurrencyCode,
			Detail:       item.Detail,
			OrgId:        item.OrgId,
			UserId:       item.UserId,
			Username:     item.Username,
			PayType:      item.PayType,
			Id:           item.Id,
		})
	}

	return response.OK(&types.TransactionRecordsListResponseData{
		Trans: items,
		Total: total,
	})
}

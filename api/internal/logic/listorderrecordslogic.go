package logic

import (
	"context"
	"encoding/json"
	"fmt"
	"knowsource/common/response"

	"knowsource/api/internal/svc"
	"knowsource/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListOrderRecordsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewListOrderRecordsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListOrderRecordsLogic {
	return &ListOrderRecordsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ListOrderRecordsLogic) ListOrderRecords(req *types.OrderRecordsListRequest) response.Response {
	uid, _ := l.ctx.Value("uid").(json.Number).Int64()
	if uid == 0 {
		return response.Fail(response.UnauthorizedCode, "Please login.")
	}

	condition := "where 1=1"
	if uid > 0 {
		condition += fmt.Sprintf(" and user_id = %d", uid)
	}
	if req.OrderType > 0 {
		condition += fmt.Sprintf(" and order_type = %d", req.OrderType)
	}
	if req.OrderNo != "" {
		condition += fmt.Sprintf(" and order_no = '%s'", req.OrderNo)
	}
	if req.BeginCreatedAt > 0 {
		condition += fmt.Sprintf(" and created_at > %d", req.BeginCreatedAt)
	}
	if req.EndCreatedAt > 0 {
		condition += fmt.Sprintf(" and created_at < %d", req.EndCreatedAt)
	}
	if req.PageSize == 0 {
		req.PageSize = 50
	}
	if req.Page == 0 {
		req.Page = 1
	}

	total, err := l.svcCtx.OrderRecordsModel.Count(l.ctx, condition)
	if err != nil {
		return response.Fail(response.ServerErrorCode, err.Error())
	}

	condition += fmt.Sprintf(" limit %d,%d", (req.Page-1)*req.PageSize, req.PageSize)

	data, err := l.svcCtx.OrderRecordsModel.FindList(l.ctx, condition)
	if err != nil {
		return response.Fail(response.ServerErrorCode, err.Error())
	}

	var orders []types.OrderRecords

	// Id            uint64 `json:"id"`
	// OrderNo       string `json:"orderNo"`
	// InstanceId    uint64 `json:"instanceId"`
	// RunningTime   uint64 `json:"runningTime"`
	// CreatedAt     uint64 `json:"createdAt"`
	// BillingMethod uint64 `json:"billingMethod"`
	// OrderAmount   uint64 `json:"orderAmount"`
	// Discount      uint64 `json:"discount"`
	// ActualAmount  uint64 `json:"actualAmount"`
	// UserId        uint64 `json:"userId"`
	// OrderType     uint64 `json:"orderType"`

	// Id            uint64 `db:"id"`             // ID
	// OrderNo       string `db:"order_no"`       // 订单号
	// InstanceId    uint64 `db:"instance_id"`    // 实例ID
	// RunningTime   uint64 `db:"running_time"`   // 运行时长
	// CreatedAt     uint64 `db:"created_at"`     // 创建时间
	// BillingMethod uint64 `db:"billing_method"` // 计费方式（1-按时、2-按天、3-按周、4-按月、0-按量）
	// OrderAmount   uint64 `db:"order_amount"`   // 订单金额
	// Discount      uint64 `db:"discount"`       // 补贴减免
	// ActualAmount  uint64 `db:"actual_amount"`  // 实收金额
	// UserId        uint64 `db:"user_id"`        // 用户ID
	// OrderType     uint64 `db:"order_type"`     // 订单类型（1-收入、2-消费）

	for _, item := range *data {
		order := types.OrderRecords{
			Id:            item.Id,
			InstanceId:    item.InstanceId,
			OrderNo:       item.OrderNo,
			OrderType:     item.OrderType,
			RunningTime:   item.RunningTime,
			BillingMethod: item.BillingMethod,
			OrderAmount:   item.OrderAmount,
			Discount:      item.Discount,
			ActualAmount:  item.ActualAmount,
			UserId:        item.UserId,
			CreatedAt:     uint64(item.CreatedAt.Unix()),
		}
		orders = append(orders, order)
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
	return response.OK(&types.OrderRecordsListResponseData{
		Orders: orders,
		Total:  total,
	})
}

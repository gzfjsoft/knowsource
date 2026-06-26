package common

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"knowsource/api/internal/svc"
	"knowsource/api/internal/types"
	"knowsource/common/response"
	"knowsource/model"
	"time"

	"github.com/zeromicro/go-zero/core/logx"
)

type UserBalancesLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUserBalancesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UserBalancesLogic {
	return &UserBalancesLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UserBalancesLogic) CheckAndUpdate(userId uint64, OrgId uint64, money int64, currencyCode string) (*model.Balances, error) {
	item, err := l.svcCtx.BalancesModel.FindOneByUserAndCurrency(l.ctx, userId, OrgId, currencyCode)
	if err == model.ErrNotFound {
		// 如果找不到，则创建一个
		item = &model.Balances{
			UserId:       userId,
			OrgId:        OrgId,
			Balance:      0,
			CurrencyCode: currencyCode,
		}

		result, err := l.svcCtx.BalancesModel.Insert(l.ctx, item)
		if err != nil {
			return nil, err
		}
		id, err := result.LastInsertId()
		if err != nil {
			return nil, err
		}
		item.BalanceId = int64(id)

	} else if err != nil {
		return item, err
	}

	item.Balance += money
	if item.Balance >= 0 {
		err = l.svcCtx.BalancesModel.UpdateBalance(l.ctx, item)
	} else {
		return nil, errors.New("balance  is not valid")
	}
	return item, err
}

func (l *UserBalancesLogic) Recharge(userId uint64, OrgId uint64, money int64, orderNo string, payType string, currencyCode string) error {
	moneyFloat := float64(money) / 100
	detail := fmt.Sprintf("%s账户：%s充值%.2f元", currencyCode, payType, moneyFloat)
	err, _ := l.Increase(userId, OrgId, money, orderNo, payType, currencyCode, detail)
	return err
}

func (l *UserBalancesLogic) Increase(userId uint64, OrgId uint64, money int64, orderNo string, payType string, currencyCode string, detail string) (error error, balance *model.Balances) {
	item, err := l.CheckAndUpdate(userId, OrgId, money, currencyCode)
	if err != nil {
		return err, nil
	}

	//transLogic := NewCreateTransactionRecordsLogic(l.ctx, l.svcCtx)
	resp := l.CreateTransactionRecordsInnerCall(&types.CreateTransactionRecordsRequest{
		TransType:    1, // 1-入金; 2-出金
		CurrencyCode: currencyCode,
		OrgId:        item.OrgId,
		UserId:       userId,
		PayType:      payType,
		Detail:       detail,
		Amount:       money,
		OrderNo:      orderNo,
		Balance:      item.Balance,
	})

	if resp.Code == response.SuccessCode {
		return nil, item
	}
	return errors.New(resp.Message), nil
}

func (l *UserBalancesLogic) Decrease(userId uint64, OrgId uint64, money int64, orderNo string, currencyCode string, detail string) (error error, balance *model.Balances) {
	item, err := l.CheckAndUpdate(userId, OrgId, -money, currencyCode)
	if err != nil {
		return err, nil
	}

	// transLogic := NewCreateTransactionRecordsLogic(l.ctx, l.svcCtx)
	resp := l.CreateTransactionRecordsInnerCall(&types.CreateTransactionRecordsRequest{
		TransType:    2, // 1-入金; 2-出金
		OrgId:        item.OrgId,
		CurrencyCode: currencyCode,
		UserId:       userId,
		PayType:      "",
		Detail:       detail,
		Amount:       money,
		OrderNo:      orderNo,
		Balance:      item.Balance,
	})

	if resp.Code == response.SuccessCode {
		return nil, item
	}
	return errors.New(resp.Message), nil
}

func (l *UserBalancesLogic) CreateTransactionRecordsInnerCall(req *types.CreateTransactionRecordsRequest) response.Response {

	if req.OrgId == 0 {
		req.OrgId = 1 // 默认充值到平台
	}

	user, err := l.svcCtx.UsersModel.FindOne(l.ctx, req.UserId)
	if err != nil {
		return response.Fail(response.UserNotExistCode, err.Error())
	}
	_, err = l.svcCtx.BalancesModel.FindOneByUserAndCurrency(l.ctx, req.UserId, req.OrgId, req.CurrencyCode)
	if err != nil {
		return response.Fail(response.UserBalanceNotExistCode, err.Error())
	}
	_, err = l.svcCtx.TransactionRecordsModel.Insert(l.ctx, &model.TransactionRecords{
		Id:           0,
		UserId:       req.UserId,
		OrgId:        req.OrgId,
		TransType:    req.TransType,
		CurrencyCode: req.CurrencyCode,
		PayType:      req.PayType,
		Detail:       req.Detail,
		OrderNo:      sql.NullString{String: req.OrderNo, Valid: req.OrderNo != ""},
		Username:     user.Username,
		Amount:       req.Amount,
		Balance:      req.Balance,
	})
	if err != nil {
		response.Fail(response.ServerErrorCode, err.Error())
	}
	return response.OK("")
}

func (l *UserBalancesLogic) ManualAdjust(userId uint64, OrgId uint64, money int64, currencyCode string, payType string, detail string) (error error, balance *model.Balances) {
	item, err := l.CheckAndUpdate(userId, OrgId, money, currencyCode)
	if err != nil {
		return err, nil
	}

	// transLogic := NewCreateTransactionRecordsLogic(l.ctx, l.svcCtx)
	resp := l.CreateTransactionRecordsInnerCall(&types.CreateTransactionRecordsRequest{
		TransType:    3, // 1-入金; 2-出金;3-其他
		OrgId:        item.OrgId,
		CurrencyCode: currencyCode,
		UserId:       userId,
		PayType:      payType,
		OrderNo:      "AJ" + time.Now().Format("20060102150405"),
		Detail:       detail,
		Amount:       money,
		Balance:      item.Balance,
	})

	if resp.Code == response.SuccessCode {
		return nil, item
	}
	return errors.New(resp.Message), nil
}

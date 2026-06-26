package logic

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"knowsource/api/internal/svc"
	"knowsource/api/internal/types"
	"knowsource/common/response"
	"knowsource/consts"
	"knowsource/model"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type AdminInstanceChargeDeletedLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewAdminInstanceChargeDeletedLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AdminInstanceChargeDeletedLogic {
	return &AdminInstanceChargeDeletedLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *AdminInstanceChargeDeletedLogic) AdminInstanceChargeDeleted() (resp *types.Response, err error) {
	resp = &types.Response{}

	sysrole, _ := l.ctx.Value("role").(string)
	if sysrole != consts.SUPER_ADMIN {
		resp.Code = response.UnauthorizedCode
		resp.Message = "没有权限"
		return resp, nil
	}

	instances, err := l.svcCtx.InstanceModel.FindAll(l.ctx, " state='deleted'")
	if err != nil {
		resp.Code = response.ServerErrorCode
		resp.Message = "获取实例失败"
		return resp, nil
	}

	for _, instance := range instances {

		///
		err := l.svcCtx.Mysql.Transact(func(session sqlx.Session) error {

			// Update user's CNY balance
			hourlyUsageModel := l.svcCtx.HourlyUsageModel.WithSession(session)
			balanceModel := l.svcCtx.BalancesModel.WithSession(session)
			transactionRecordsModel := l.svcCtx.TransactionRecordsModel.WithSession(session)

			balance, err := balanceModel.FindOneByUserAndCurrency(l.ctx, instance.UserId, instance.OrgId, "CNY")
			if err != nil {
				return fmt.Errorf("failed to get user balance: %v", err)
			}

			usages, err := hourlyUsageModel.FindUncharged(l.ctx, instance.OrgId, instance.UserId, instance.InstanceId)
			if err != nil {
				return fmt.Errorf("failed to get unbilled usages: %v", err)
			}

			for _, usage := range usages {

				balance.Balance -= int64(usage.Fee)

				transRecord := &model.TransactionRecords{
					OrgId:        instance.OrgId,
					UserId:       instance.UserId,
					Amount:       int64(usage.Fee),
					OrderNo:      sql.NullString{String: instance.Name, Valid: true},
					TransType:    2,
					PayType:      "租用费用",
					CurrencyCode: "CNY",
					Detail:       "使用实例" + instance.Name + "的费用",
					Username:     instance.Name,
					Balance:      int64(balance.Balance),
					CreatedAt:    time.Now(),
				}

				_, err = transactionRecordsModel.Insert(l.ctx, transRecord)
				if err != nil {
					return fmt.Errorf("failed to insert transaction record: %v", err)
				}

				usage.IsCharged = 1
				err = hourlyUsageModel.Update(l.ctx, usage)
				if err != nil {
					return fmt.Errorf("failed to update hourly usage: %v", err)
				}

			}

			err = balanceModel.Update(l.ctx, balance)
			if err != nil {
				return fmt.Errorf("failed to update user balance: %v", err)
			}

			return nil
		})

		if err != nil {
			logx.Errorf("!!! Failed to update user balance: %v", err)
		}

	}

	if err != nil {
		resp.Code = response.ServerErrorCode
		resp.Message = "计算实例失败"
		resp.Info = err.Error()
		return resp, nil
	}

	resp.Code = response.SuccessCode
	resp.Message = "获取实例成功"
	return resp, nil

}

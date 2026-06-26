package logic

import (
	"context"
	"encoding/json"
	"knowsource/common/response"

	"knowsource/api/internal/svc"
	"knowsource/api/internal/types"
	"knowsource/model"

	"github.com/zeromicro/go-zero/core/logx"
)

type UserLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUserLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UserLogic {
	return &UserLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UserLogic) User() (resp response.Response) {
	uid, _ := l.ctx.Value("uid").(json.Number).Int64()
	user, err := l.svcCtx.UsersModel.FindOne(l.ctx, uint64(uid))
	if err != nil {
		if err == model.ErrNotFound {
			return response.Fail(response.UserNotExistCode, "用户不存在")
		}
		return response.Fail(response.ServerErrorCode, err.Error())
	}

	// Fetch user's organizations
	modelOrgs, err := l.svcCtx.OrgsUsersModel.FindAllByUserId(l.ctx, user.UserId)
	if err != nil {
		resp.Code = response.ServerErrorCode
		resp.Message = "Failed to fetch user organizations"
		return resp
	}

	// Convert model.Organizations to types.Org
	orgs := make([]types.Org, len(*modelOrgs))
	for i, org := range *modelOrgs {
		orgs[i] = types.Org{
			OrgId:   org.OrgId,
			OrgName: org.OrgName,
			Role:    org.Role,
		}
	}

	// balance, err := l.svcCtx.BalancesModel.FindOneByUserAndCurrency(l.ctx, user.UserId, 0, "CNY")
	// if err != nil {
	// 	resp.Code = response.ServerErrorCode
	// 	resp.Message = "Failed to fetch user balance"
	// 	return resp
	// }

	return response.OK(&types.UserResponseData{
		UserId:          user.UserId,
		Username:        user.Username,
		Nickname:        user.Nickname,
		Email:           user.Email,
		Phone:           user.Phone,
		HeadUrl:         user.HeadUrl,
		SysRole:         user.SysRole,
		CreatedAt:       uint64(user.CreatedAt.Unix()),
		IsPhoneVerified: int64(user.IsPhoneVerified),
		IsEmailVerified: int64(user.IsEmailVerified),
		Uuid:            user.Uuid,
		Avatar:          user.Avatar,
		LoginedAt:       uint64(user.LoginedAt.Unix()),

		IsDeleted: int64(user.IsDeleted),
		Orgs:      orgs,
	})
}

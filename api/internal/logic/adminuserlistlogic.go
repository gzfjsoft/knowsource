package logic

import (
	"context"
	"encoding/json"
	"fmt"

	"knowsource/api/internal/svc"
	"knowsource/api/internal/types"
	"knowsource/api/internal/utils"
	"knowsource/common/response"

	"github.com/zeromicro/go-zero/core/logx"
)

type AdminUserListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewAdminUserListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AdminUserListLogic {
	return &AdminUserListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *AdminUserListLogic) AdminUserList(req *types.AdminUserListRequest) (resp *types.AdminUserListResponse, err error) {
	resp = &types.AdminUserListResponse{
		Response: types.Response{
			Code:    response.SuccessCode,
			Message: "Success",
		},
	}

	// 管理员权限验证 start
	uid, _ := l.ctx.Value("uid").(json.Number).Int64()
	email, _ := l.ctx.Value("email").(string)
	if !utils.IsSuperAdminRoleFromContext(l.ctx) {
		role, _ := utils.GetRoleFromContext(l.ctx)
		resp.Code = response.UnauthorizedCode
		resp.Message = fmt.Sprintf("权限不足，需要superadmin，你的权限是 %s.", role)
		resp.Info = role
		return resp, nil
	}

	logx.Info("JWT uid=", uid, " Name=", email)

	users, total, err := l.svcCtx.UsersModel.FindUsers(l.ctx, req.Username, req.Email, req.Phone, req.Page, req.PageSize)
	if err != nil {
		resp.Code = response.ServerErrorCode
		resp.Message = "Failed to fetch users: " + err.Error()
		return resp, err
	}

	// balancemap := make(map[uint64]int64)
	// condition := ""
	// for _, user := range users {
	// 	balancemap[user.UserId] = 0
	// 	condition += fmt.Sprintf("%d,", user.UserId)
	// }

	// balances, err := l.svcCtx.BalancesModel.FindList(l.ctx, "where currency_code = 'CNY' and user_id in ("+condition+")")
	// if err != nil {
	// 	resp.Code = response.ServerErrorCode
	// 	resp.Message = "Failed to fetch balances: " + err.Error()
	// 	return resp, err
	// }

	// for _, balance := range *balances {
	// 	balancemap[balance.UserId] = balance.Balance
	// }

	userList := make([]types.UserResponseData, len(users))
	for i, user := range users {
		userList[i] = types.UserResponseData{
			UserId:   user.UserId,
			Username: user.Username,
			Nickname: user.Nickname,
			Phone:    user.Phone,
			Email:    user.Email,
			HeadUrl:  user.HeadUrl,
			SysRole:  user.SysRole,

			IsDeleted:       int64(user.IsDeleted),
			IsPhoneVerified: int64(user.IsPhoneVerified),
			IsEmailVerified: int64(user.IsEmailVerified),
			Uuid:            user.Uuid,
			Avatar:          user.Avatar,
			LoginedAt:       uint64(user.LoginedAt.Unix()),
			CreatedAt:       uint64(user.CreatedAt.Unix()),

			//TODO : no need to add org, because it will be fetched in the frontend
			// Note: Orgs field is not populated here. If needed, you should fetch and populate it separately.
		}
	}

	resp.Data = types.AdminUserListResponseData{
		Users: userList,
		Total: total,
	}

	return resp, nil
}

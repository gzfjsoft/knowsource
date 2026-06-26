package logic

import (
	"context"
	"encoding/json"
	"fmt"

	"knowsource/api/internal/svc"
	"knowsource/api/internal/types"
	"knowsource/common/response"
	"knowsource/consts"

	"github.com/zeromicro/go-zero/core/logx"
)

type AdminUserDeleteLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewAdminUserDeleteLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AdminUserDeleteLogic {
	return &AdminUserDeleteLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *AdminUserDeleteLogic) AdminUserDelete(req *types.AdminUserDeleteRequest) (resp *types.Response, err error) {
	resp = &types.Response{
		Code:    response.SuccessCode,
		Message: "Success",
	}

	uid, _ := l.ctx.Value("uid").(json.Number).Int64()
	email, _ := l.ctx.Value("email").(string)
	logx.Info("JWT uid=", uid, " Name=", email)

	// 管理员权限验证 start
	role, _ := l.ctx.Value("role").(string)

	if role != consts.SUPER_ADMIN {
		resp.Code = response.UnauthorizedCode
		resp.Message = fmt.Sprintf("权限不足，需要superadmin，你的权限是%s.", role)
		resp.Info = role
		return resp, nil
	}
	// 管理员权限验证 end

	for _, userId := range req.UserIds {

		if userId == 0 {
			continue
		}

		// var user *model.Users

		// user, err = l.svcCtx.UsersModel.FindOneWithDelete(l.ctx, userId)

		// if err != nil {
		// 	if err != sqlx.ErrNotFound {

		// 		resp.Code = response.ServerErrorCode
		// 		resp.Message = "服务器错误 "
		// 		resp.Info = err.Error()
		// 		return resp, err
		// 	}
		// }

		// if user.IsDeleted > 0 {
		// 	return resp, nil
		// }

		balances, err := l.svcCtx.BalancesModel.FindList(l.ctx, fmt.Sprintf("where user_id = %d", userId))
		if err != nil {
			resp.Code = response.ServerErrorCode
			resp.Message = "删除用户失败: 查询余额失败" + err.Error()
			resp.Info = fmt.Sprintf("删除用户失败，查询余额失败 %d", userId)
			return resp, err
		}

		for _, balance := range *balances {
			if balance.Balance > 0 {
				org, err := l.svcCtx.OrganizationModel.FindOne(l.ctx, balance.OrgId)
				if err != nil {
					resp.Code = response.ServerErrorCode
					resp.Message = "删除用户失败: 查询组织失败"
					resp.Info = fmt.Sprintf("删除用户失败，查询组织失败 %d", balance.OrgId)
					return resp, err
				}
				resp.Code = response.ServerErrorCode
				resp.Message = fmt.Sprintf("删除用户失败: 余额大于0 (%s)", org.OrgName)
				resp.Info = fmt.Sprintf("删除用户失败，余额大于0 %d %s", userId, org.OrgName)
				return resp, nil
			}

		}

		// 删除余额
		err = l.svcCtx.BalancesModel.DeleteByUserId(l.ctx, userId)
		if err != nil {
			resp.Code = response.ServerErrorCode
			resp.Message = "删除用户失败: 删除余额失败"
			resp.Info = fmt.Sprintf("删除用户失败，删除余额失败 %d", userId)
			return resp, err
		}
		// 删除组织用户

		err = l.svcCtx.OrgsUsersModel.DeleteByUserId(l.ctx, userId)
		if err != nil {
			resp.Code = response.ServerErrorCode
			resp.Message = "删除用户失败: 删除组织用户失败"
			resp.Info = fmt.Sprintf("删除用户失败，删除组织用户失败 %d", userId)
			return resp, err
		}
		// 删除用户角色
		err = l.svcCtx.UserRolesModel.DeleteByUserId(l.ctx, userId)
		if err != nil {
			resp.Code = response.ServerErrorCode
			resp.Message = "删除用户失败: 删除用户角色失败"
			resp.Info = fmt.Sprintf("删除用户失败，删除用户角色失败 %d", userId)
			return resp, err
		}

		err = l.svcCtx.UsersModel.Delete(l.ctx, userId)
		if err != nil {
			resp.Code = response.ServerErrorCode
			resp.Message = "删除用户失败: 删除用户失败"
			resp.Info = err.Error()
			return resp, err
		}

		// 删除用户
		// 删除用户
		err = l.svcCtx.UsersModel.DeleteWithDelete(l.ctx, userId)
		if err != nil {
			resp.Code = response.ServerErrorCode
			resp.Message = "删除用户失败: "
			resp.Info = err.Error()
			// return resp, err
		}

	}
	resp.Message = "删除用户成功"
	resp.Code = response.SuccessCode

	return resp, nil
}

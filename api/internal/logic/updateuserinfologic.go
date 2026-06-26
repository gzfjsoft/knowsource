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

type UpdateUserInfoLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUpdateUserInfoLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateUserInfoLogic {
	return &UpdateUserInfoLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

//user side

func (l *UpdateUserInfoLogic) UpdateUserInfo(req *types.UpdateUserInfoRequest) (resp *types.Response, err error) {

	resp = new(types.Response)

	uid, err := l.ctx.Value("uid").(json.Number).Int64()
	if err != nil {
		resp.Code = response.UnauthorizedCode
		resp.Message = "Invalid user ID"
		return resp, nil
	}
	userId := uint64(uid)

	user, err := l.svcCtx.UsersModel.FindOne(l.ctx, userId)
	if err != nil {
		if err == model.ErrNotFound {
			resp.Code = 404
			resp.Message = "User not found"
			resp.Info = err.Error()
			return resp, nil
		} else {
			resp.Code = response.ServerErrorCode
			resp.Message = "更新用户信息失败"
			resp.Info = err.Error()
			return resp, nil
		}

	}

	if req.Email != "" {
		user.Email = req.Email
		user.IsEmailVerified = 0
	}
	if req.Username != "" {
		user.Username = req.Username
	}
	if req.Nickname != "" {
		user.Nickname = req.Nickname
	}

	if req.Phone != "" {
		user.Phone = req.Phone
		user.IsPhoneVerified = 0
	}

	err = l.svcCtx.UsersModel.UpdateInfo(l.ctx, user)
	if err != nil {
		resp.Code = response.ServerErrorCode
		resp.Message = "更新用户信息失败"
		resp.Info = err.Error()
		return resp, nil

	}

	resp.Code = response.SuccessCode
	resp.Message = "success"
	return resp, nil
}

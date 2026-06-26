package logic

import (
	"context"
	"encoding/json"
	"time"

	"knowsource/api/internal/svc"
	"knowsource/api/internal/types"
	"knowsource/common/jwtx"
	"knowsource/common/response"

	"github.com/zeromicro/go-zero/core/logx"
)

type KeepLoginLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewKeepLoginLogic(ctx context.Context, svcCtx *svc.ServiceContext) *KeepLoginLogic {
	return &KeepLoginLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *KeepLoginLogic) KeepLogin() (resp *types.LoginResponse, err error) {

	uid, err := l.ctx.Value("uid").(json.Number).Int64()
	if err != nil {
		return &types.LoginResponse{
			Response: types.Response{
				Code:    response.UnauthorizedCode,
				Message: "Invalid user ID",
			},
		}, nil
	}
	userId := uint64(uid)

	clientId, _ := l.ctx.Value("clientId").(string)
	if clientId == "" {
		return &types.LoginResponse{
			Response: types.Response{
				Code:    response.UnauthorizedCode,
				Message: "clientId不能为空，请重新登录",
			},
		}, nil
	}

	user, err := l.svcCtx.UsersModel.FindOne(l.ctx, userId)
	if err != nil {
		return &types.LoginResponse{
			Response: types.Response{
				Code:    response.UnauthorizedCode,
				Message: "Invalid user ID",
			},
		}, nil
	}

	now := time.Now().Unix()
	accessExpire := l.svcCtx.Config.Auth.AccessExpire

	token, err := jwtx.GetToken(l.svcCtx.Config.Auth.AccessSecret, now, accessExpire, clientId, (user.UserId), user.Email, user.SysRole, user.Uuid)
	if err != nil {
		return &types.LoginResponse{
			Response: types.Response{
				Code:    response.UnauthorizedCode,
				Message: "Invalid user ID",
			},
		}, nil
	}

	return &types.LoginResponse{
		Response: types.Response{
			Code:    response.SuccessCode,
			Message: "Keep login successful",
		},
		Data: &types.LoginResponseData{
			AccessToken:  token,
			AccessExpire: now + accessExpire,
			Uuid:         user.Uuid,
			Avatar:       user.Avatar,
			Username:     user.Username,
			Nickname:     user.Nickname,
			SysRole:      user.SysRole,
		},
	}, nil
}

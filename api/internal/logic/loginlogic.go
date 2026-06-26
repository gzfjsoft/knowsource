package logic

import (
	"context"
	"knowsource/common/response"
	"time"

	"knowsource/api/internal/svc"
	"knowsource/api/internal/types"
	"knowsource/api/internal/utils"
	"knowsource/common/cryptx"
	"knowsource/common/jwtx"
	"knowsource/model"

	"github.com/zeromicro/go-zero/core/logx"
)

type LoginLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewLoginLogic(ctx context.Context, svcCtx *svc.ServiceContext) *LoginLogic {
	return &LoginLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *LoginLogic) Login(req *types.LoginRequest) (res response.Response) {
	if req.ClientId == "" {
		return response.Fail(response.ParameterErrorCode, "clientId不能为空")
	}

	if !utils.VerifyCaptcha(req.CaptchaId, req.Captcha) {
		logx.Infof("验证码错误: %s, %s", req.CaptchaId, req.Captcha)
		return response.Fail(response.WrongCaptchaCode, "验证码错误")
	}

	var user *model.Users
	var err error

	if req.Username != "" {
		user, err = l.svcCtx.UsersModel.FindOneByUsername(l.ctx, req.Username)
		if err != nil {
			if err == model.ErrNotFound {
				return response.Fail(100, "用户不存在")
			}
			return response.Error(err.Error())
		}
	} else {
		user, err = l.svcCtx.UsersModel.FindOneByEmail(l.ctx, req.Email)
		if err != nil {
			if err == model.ErrNotFound {
				return response.Fail(100, "用户不存在")
			}
			return response.Error(err.Error())
		}

	}

	if user.IsEmailVerified == 0 {
		code := generateVerificationCode()
		err = sendVerificationEmail(l.svcCtx, user.Email, "register", code)
		if err != nil {
			logx.Error("登陆验证码发送失败: %s", err.Error())
		}
		if err == nil {
			err = storeVerificationData(l.svcCtx, user.Email, code, "")
			if err != nil {
				logx.Error("登陆验证码存储失败: %s", err.Error())
			}
		}

		return response.FailWithInfo(458, "你的 email 还没有被校验，请检查你的邮箱", "")
	}

	password := cryptx.PasswordEncrypt(l.svcCtx.Config.Salt, req.Password)
	if user.PasswordHash != password {
		return response.Fail(response.ForbiddenCode, "密码错误")
	}

	now := time.Now().Unix()
	accessExpire := l.svcCtx.Config.Auth.AccessExpire
	token, err := jwtx.GetToken(l.svcCtx.Config.Auth.AccessSecret, now, accessExpire, req.ClientId, (user.UserId), user.Email, user.SysRole, user.Uuid)
	if err != nil {
		return response.Error(err.Error())
	}

	orgs, err := l.svcCtx.OrgsUsersModel.FindAllByUserId(l.ctx, user.UserId)
	if err != nil {
		return response.Error(err.Error())
	}

	// Convert []*model.Org to []types.Org
	typesOrgs := make([]types.Org, len(*orgs))
	for i, org := range *orgs {
		typesOrgs[i] = types.Org{
			OrgId:   org.OrgId,
			OrgName: org.OrgName,
			Role:    org.Role,
			// Add other fields as needed
		}
	}

	//
	_, err = l.svcCtx.UsersLoginLogModel.Insert(l.ctx, &model.UsersLoginLog{
		UserId: user.UserId,
		Uuid:   user.Uuid,
		// Ip:        req.Ip,
		LoginedAt: time.Now(),
	})
	if err != nil {
		logx.Error("用户登录日志记录失败: %s", err.Error())
	}

	return response.OK(&types.LoginResponseData{
		AccessToken:  token,
		AccessExpire: now + accessExpire,
		Uuid:         user.Uuid,
		Avatar:       user.Avatar,
		Username:     user.Username,
		Nickname:     user.Nickname,
		SysRole:      user.SysRole,
		Orgs:         typesOrgs,
	})
}

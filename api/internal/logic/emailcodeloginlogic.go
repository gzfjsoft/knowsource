package logic

import (
	"context"
	"time"

	"knowsource/api/internal/svc"
	"knowsource/api/internal/types"
	"knowsource/common/jwtx"
	"knowsource/common/response"
	"knowsource/model"

	"github.com/zeromicro/go-zero/core/logx"
)

type EmailcodeloginLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewEmailcodeloginLogic(ctx context.Context, svcCtx *svc.ServiceContext) *EmailcodeloginLogic {
	return &EmailcodeloginLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *EmailcodeloginLogic) Emailcodelogin(req *types.EmailcodeloginRequest) (*types.LoginResponse, error) {
	// Helper function to generate error response
	errorResponse := func(code int64, message string, info string) *types.LoginResponse {
		return &types.LoginResponse{
			Response: types.Response{
				Code:    code,
				Message: message,
				Info:    info,
			},
		}
	}

	if req.ClientId == "" {
		return errorResponse(response.ParameterErrorCode, "clientId不能为空", ""), nil
	}

	// Find user by email
	user, err := l.svcCtx.UsersModel.FindOneByEmail(l.ctx, req.Email)
	if err != nil {
		if err == model.ErrNotFound {
			logx.Error("用户不存在")
			return errorResponse(404, "User not found", "The provided email is not associated with any user"), nil
		}
		return errorResponse(response.ServerErrorCode, "服务器错误，请联系客服", "(FindOneByEmail)"+err.Error()), nil
	}

	// Verify the provided code
	verificationCode, err := l.svcCtx.VerificationCodesModel.FindOneByEmail(l.ctx, req.Email)
	if err != nil {
		if err == model.ErrNotFound {
			logx.Error("邮件验证码不存在或已过期")
			return errorResponse(404, "邮件验证码不存在或已过期", "Please request a new verification code"), nil
		}
		return errorResponse(response.ServerErrorCode, "服务器错误，请联系客服", "(FindOneByEmail)"+err.Error()), nil
	}

	if verificationCode.Code != req.Code {
		logx.Error("邮件验证码错误", verificationCode.Code, req.Code)
		return errorResponse(response.ParameterErrorCode, "邮件验证码错误", "The provided code does not match"), nil
	}

	// Check if the code has expired (assuming 5 minutes expiration)
	if time.Now().After(verificationCode.ExpirationTime) {
		logx.Error("邮件验证码已过期")
		return errorResponse(response.ParameterErrorCode, "Verification code expired", "Please request a new verification code"), nil
	}

	// Generate JWT token
	now := time.Now().Unix()
	accessExpire := l.svcCtx.Config.Auth.AccessExpire
	token, err := jwtx.GetToken(l.svcCtx.Config.Auth.AccessSecret, now, accessExpire, req.ClientId, user.UserId, user.Email, user.SysRole, user.Uuid)
	if err != nil {
		return errorResponse(response.ServerErrorCode, "Failed to generate token", err.Error()), nil
	}

	// Delete the used verification code
	err = l.svcCtx.VerificationCodesModel.Delete(l.ctx, verificationCode.Id)
	if err != nil {
		logx.Error("Failed to delete verification code:", err)
	}

	// Fetch user's organizations
	modelOrgs, err := l.svcCtx.OrgsUsersModel.FindAllByUserId(l.ctx, user.UserId)
	if err != nil {
		return errorResponse(response.ServerErrorCode, "Failed to fetch user organizations", err.Error()), nil
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

	// Successful response
	return &types.LoginResponse{
		Response: types.Response{
			Code:    response.SuccessCode,
			Message: "Login successful",
			Info:    "User authenticated successfully",
		},
		Data: &types.LoginResponseData{
			AccessToken:  token,
			AccessExpire: now + accessExpire,
			Uuid:         user.Uuid,
			Avatar:       user.Avatar,
			Username:     user.Username,
			Nickname:     user.Nickname,
			SysRole:      user.SysRole,
			Orgs:         orgs,
		},
	}, nil
}

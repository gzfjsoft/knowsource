package knowsource

import (
	"context"
	"strings"

	"knowsource/api/internal/svc"
	"knowsource/api/internal/types"
	"knowsource/common/response"
	"knowsource/model"

	"github.com/zeromicro/go-zero/core/logx"
)

type KnowsourcePasswordResetSendCodeLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewKnowsourcePasswordResetSendCodeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *KnowsourcePasswordResetSendCodeLogic {
	return &KnowsourcePasswordResetSendCodeLogic{Logger: logx.WithContext(ctx), ctx: ctx, svcCtx: svcCtx}
}

func (l *KnowsourcePasswordResetSendCodeLogic) KnowsourcePasswordResetSendCode(req *types.KnowsourcePasswordResetSendCodeRequest, clientIP string) (resp *types.Response, err error) {
	if req == nil {
		return &types.Response{Code: response.ParameterErrorCode, Message: "参数不能为空"}, nil
	}
	clientId := strings.TrimSpace(req.ClientId)
	ch := strings.ToLower(strings.TrimSpace(req.Channel))
	if clientId == "" || (ch != "email" && ch != "phone") {
		return &types.Response{Code: response.ParameterErrorCode, Message: "clientId 与 channel（email|phone）必填"}, nil
	}

	var emp *model.FrEmp
	var e error
	if ch == "email" {
		em := strings.TrimSpace(strings.ToLower(req.Email))
		if em == "" {
			return &types.Response{Code: response.ParameterErrorCode, Message: "请填写 email"}, nil
		}
		emp, e = l.svcCtx.FrEmpModel.FindOneByClientIdEmail(l.ctx, clientId, em)
	} else {
		ph := strings.TrimSpace(req.Phone)
		if ph == "" {
			return &types.Response{Code: response.ParameterErrorCode, Message: "请填写 phone"}, nil
		}
		emp, e = l.svcCtx.FrEmpModel.FindOneByClientIdMobile(l.ctx, clientId, ph)
	}
	if e != nil {
		if e == model.ErrNotFound {
			return &types.Response{Code: response.UserNotExistCode, Message: "未找到对应员工"}, nil
		}
		return &types.Response{Code: response.ServerErrorCode, Message: "Database error", Info: e.Error()}, nil
	}

	resetKey := clientId + "|" + emp.FempCode
	code := KnowsourceRandomDigitCode()
	if e := KnowsourceStoreVerificationCode(l.ctx, l.svcCtx, targetKnowsourcePwdReset, resetKey, code, clientIP); e != nil {
		return &types.Response{Code: response.ServerErrorCode, Message: "写入验证码失败", Info: e.Error()}, nil
	}

	if ch == "email" {
		em := strings.TrimSpace(strings.ToLower(req.Email))
		subject := "【知识库】重置密码验证码"
		body := "<p>您的验证码为：<strong>" + code + "</strong></p><p>15 分钟内有效。</p>"
		if e := KnowsourceSendSimpleMail(l.svcCtx, em, subject, body); e != nil {
			return &types.Response{Code: response.ServerErrorCode, Message: "发送邮件失败", Info: e.Error()}, nil
		}
	} else {
		ph := strings.TrimSpace(req.Phone)
		if e := KnowsourceSendSMSCode(l.svcCtx, ph, code); e != nil {
			return &types.Response{Code: response.ServerErrorCode, Message: "发送短信失败", Info: e.Error()}, nil
		}
	}

	return &types.Response{Code: response.SuccessCode, Message: "验证码已发送"}, nil
}

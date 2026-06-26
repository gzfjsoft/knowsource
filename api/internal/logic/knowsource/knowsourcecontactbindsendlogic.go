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

type KnowsourceContactBindSendLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewKnowsourceContactBindSendLogic(ctx context.Context, svcCtx *svc.ServiceContext) *KnowsourceContactBindSendLogic {
	return &KnowsourceContactBindSendLogic{Logger: logx.WithContext(ctx), ctx: ctx, svcCtx: svcCtx}
}

func (l *KnowsourceContactBindSendLogic) KnowsourceContactBindSend(req *types.KnowsourceContactBindSendRequest, clientIP string) (resp *types.Response, err error) {
	clientId, _ := l.ctx.Value("clientId").(string)
	empCode, _ := l.ctx.Value("empCode").(string)
	clientId = strings.TrimSpace(clientId)
	empCode = strings.TrimSpace(empCode)
	if clientId == "" || empCode == "" {
		return &types.Response{Code: response.UnauthorizedCode, Message: "未登录或缺少身份信息"}, nil
	}
	if req == nil {
		return &types.Response{Code: response.ParameterErrorCode, Message: "参数不能为空"}, nil
	}
	ch := strings.ToLower(strings.TrimSpace(req.Channel))
	if ch != "email" && ch != "phone" {
		return &types.Response{Code: response.ParameterErrorCode, Message: "channel 须为 email 或 phone"}, nil
	}

	if ch == "email" {
		em := strings.TrimSpace(strings.ToLower(req.NewEmail))
		if em == "" || !strings.Contains(em, "@") {
			return &types.Response{Code: response.ParameterErrorCode, Message: "请填写有效邮箱"}, nil
		}
		if o, e := l.svcCtx.FrEmpModel.FindOneByClientIdEmail(l.ctx, clientId, em); e == nil && o != nil && o.FempCode != empCode {
			return &types.Response{Code: response.ConflictCode, Message: "该邮箱已被其他员工使用"}, nil
		} else if e != nil && e != model.ErrNotFound {
			return &types.Response{Code: response.ServerErrorCode, Message: "Database error", Info: e.Error()}, nil
		}
		code := KnowsourceRandomDigitCode()
		tv := clientId + "|" + empCode + "|" + em
		if e := KnowsourceStoreVerificationCode(l.ctx, l.svcCtx, targetKnowsourceBindEmail, tv, code, clientIP); e != nil {
			return &types.Response{Code: response.ServerErrorCode, Message: "写入验证码失败", Info: e.Error()}, nil
		}
		subject := "【知源智库 AI】绑定邮箱验证码"
		body := "<p>您的验证码为：<strong>" + code + "</strong></p><p>15 分钟内有效。</p>"
		if e := KnowsourceSendSimpleMail(l.svcCtx, em, subject, body); e != nil {
			return &types.Response{Code: response.ServerErrorCode, Message: "发送邮件失败", Info: e.Error()}, nil
		}
		return &types.Response{Code: response.SuccessCode, Message: "验证码已发送至邮箱"}, nil
	}

	ph := strings.TrimSpace(req.NewPhone)
	if ph == "" {
		return &types.Response{Code: response.ParameterErrorCode, Message: "请填写手机号"}, nil
	}
	if o, e := l.svcCtx.FrEmpModel.FindOneByClientIdMobile(l.ctx, clientId, ph); e == nil && o != nil && o.FempCode != empCode {
		return &types.Response{Code: response.ConflictCode, Message: "该手机号已被其他员工使用"}, nil
	} else if e != nil && e != model.ErrNotFound {
		return &types.Response{Code: response.ServerErrorCode, Message: "Database error", Info: e.Error()}, nil
	}
	code := KnowsourceRandomDigitCode()
	tv := clientId + "|" + empCode + "|" + ph
	if e := KnowsourceStoreVerificationCode(l.ctx, l.svcCtx, targetKnowsourceBindPhone, tv, code, clientIP); e != nil {
		return &types.Response{Code: response.ServerErrorCode, Message: "写入验证码失败", Info: e.Error()}, nil
	}
	if e := KnowsourceSendSMSCode(l.svcCtx, ph, code); e != nil {
		return &types.Response{Code: response.ServerErrorCode, Message: "发送短信失败", Info: e.Error()}, nil
	}
	return &types.Response{Code: response.SuccessCode, Message: "验证码已发送至手机"}, nil
}

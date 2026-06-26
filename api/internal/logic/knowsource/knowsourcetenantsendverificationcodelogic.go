package knowsource

import (
	"context"
	"fmt"
	"strings"

	"knowsource/api/internal/superadmin"
	"knowsource/api/internal/svc"
	"knowsource/api/internal/types"
	"knowsource/common/response"
	"knowsource/model"

	"github.com/zeromicro/go-zero/core/logx"
)

type KnowsourceTenantSendVerificationCodeLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewKnowsourceTenantSendVerificationCodeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *KnowsourceTenantSendVerificationCodeLogic {
	return &KnowsourceTenantSendVerificationCodeLogic{Logger: logx.WithContext(ctx), ctx: ctx, svcCtx: svcCtx}
}

func (l *KnowsourceTenantSendVerificationCodeLogic) KnowsourceTenantSendVerificationCode(req *types.KnowsourceTenantSendVerificationCodeRequest) (resp *types.Response, err error) {
	if req == nil || strings.TrimSpace(req.ClientId) == "" {
		return &types.Response{Code: response.ParameterErrorCode, Message: "clientId 不能为空"}, nil
	}

	clientId := strings.TrimSpace(req.ClientId)

	// 查找租户信息
	row, e := l.svcCtx.ClientModel.FindOneByClientId(l.ctx, clientId)
	if e != nil {
		if e == model.ErrNotFound {
			return &types.Response{Code: response.NotFoundCode, Message: "租户不存在"}, nil
		}
		return &types.Response{Code: response.ServerErrorCode, Message: "Database error", Info: e.Error()}, nil
	}

	// 检查租户状态
	if row.Status != 0 {
		return &types.Response{Code: response.SuccessCode, Message: "该租户已验证或无需重复验证"}, nil
	}

	// 获取租户邮箱
	email := strings.TrimSpace(row.OwnerEmail)
	if email == "" {
		return &types.Response{Code: response.ParameterErrorCode, Message: "租户邮箱未设置"}, nil
	}

	// 生成验证码
	code := KnowsourceRandomDigitCode()

	// 存储验证码
	if e := KnowsourceStoreVerificationCode(l.ctx, l.svcCtx, targetKnowsourceTenantVerify, clientId, code, ""); e != nil {
		return &types.Response{Code: response.ServerErrorCode, Message: "存储验证码失败", Info: e.Error()}, nil
	}

	// 发送邮件
	username := superadmin.SuperadminEmpCode()
	subject := "【知源智库 AI】租户邮箱验证"
	body := fmt.Sprintf(`<div style="font-family: Arial, sans-serif; max-width: 600px; margin: 0 auto;">
<h2 style="color: #333;">租户邮箱验证</h2>
<p>尊敬的用户：</p>
<p>您正在进行知源智库 AI 租户邮箱验证，账户信息如下：</p>
<ul>
<li><strong>企业账户名</strong>：<code>%s</code></li>
<li><strong>用户名</strong>：<code>%s</code></li>
</ul>
<p>您的验证码为：</p>
<div style="background-color: #f5f5f5; padding: 20px; margin: 20px 0; border-radius: 5px;">
<p style="font-size: 24px; font-weight: bold; text-align: center; color: #333;">%s</p>
</div>
<p>验证码有效期为 15 分钟，请及时使用。</p>
<p>如果您没有发起此操作，请忽略此邮件。</p>
<p>此致<br>知源智库 AI 团队</p>
</div>`, clientId, username, code)

	if e := KnowsourceSendSimpleMail(l.svcCtx, email, subject, body); e != nil {
		return &types.Response{Code: response.ServerErrorCode, Message: "发送邮件失败", Info: e.Error()}, nil
	}

	return &types.Response{Code: response.SuccessCode, Message: "验证码已发送，请查收邮箱"}, nil
}

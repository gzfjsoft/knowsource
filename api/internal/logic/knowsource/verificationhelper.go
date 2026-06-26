package knowsource

import (
	"context"
	"crypto/rand"
	"fmt"
	"math/big"
	"strings"
	"time"

	"knowsource/api/internal/svc"
	"knowsource/model"

	"github.com/zeromicro/go-zero/core/logx"
	"gopkg.in/gomail.v2"
)

// KnowsourceRandomDigitCode 6 位数字验证码
func KnowsourceRandomDigitCode() string {
	n, err := rand.Int(rand.Reader, big.NewInt(900000))
	if err != nil {
		return fmt.Sprintf("%06d", time.Now().UnixNano()%900000+100000)
	}
	return fmt.Sprintf("%06d", n.Int64()+100000)
}

// KnowsourceStoreVerificationCode 写入 verification_codes
func KnowsourceStoreVerificationCode(ctx context.Context, svcCtx *svc.ServiceContext, targetType, targetValue, code, ip string) error {
	_, err := svcCtx.VerificationCodesModel.Insert(ctx, &model.VerificationCodes{
		TargetType:     targetType,
		TargetValue:    targetValue,
		Code:           code,
		Ip:             ip,
		ExpirationTime: time.Now().Add(15 * time.Minute),
	})
	return err
}

// KnowsourceSendSimpleMail 使用全局 Mail 配置发送 HTML 邮件
func KnowsourceSendSimpleMail(svcCtx *svc.ServiceContext, to, subject, htmlBody string) error {
	to = strings.TrimSpace(to)
	if to == "" {
		return fmt.Errorf("收件邮箱为空")
	}
	acc := strings.TrimSpace(svcCtx.Config.Mail.MailAccount)
	host := strings.TrimSpace(svcCtx.Config.Mail.MailHost)
	pass := svcCtx.Config.Mail.MailPass
	port := svcCtx.Config.Mail.MailPort
	if acc == "" || host == "" || port <= 0 {
		return fmt.Errorf("未配置 Mail（MailAccount/MailHost/MailPort）")
	}
	m := gomail.NewMessage()
	m.SetHeader("From", acc)
	m.SetHeader("To", to)
	m.SetHeader("Subject", subject)
	m.SetBody("text/html", htmlBody)
	d := gomail.NewDialer(host, port, acc, pass)
	// 163 SMTP 走 465 时通常需要 SSL；也可通过配置显式开启
	if svcCtx.Config.Mail.MailSSL || port == 465 {
		d.SSL = true
	}
	if err := d.DialAndSend(m); err != nil {
		logx.Errorf("KnowsourceSendSimpleMail failed: %v", err)
		return err
	}
	return nil
}

// KnowsourceVerifyStoredCode 校验 verification_codes 中最新一条记录
func KnowsourceVerifyStoredCode(ctx context.Context, svcCtx *svc.ServiceContext, targetType, targetValue, code string) error {
	code = strings.TrimSpace(code)
	if code == "" {
		return fmt.Errorf("验证码不能为空")
	}
	row, err := svcCtx.VerificationCodesModel.FindLatestByTypeAndValue(ctx, targetType, targetValue)
	if err != nil {
		if err == model.ErrNotFound {
			return fmt.Errorf("验证码不存在或已失效")
		}
		return err
	}
	if row.Code != code {
		return fmt.Errorf("验证码错误")
	}
	if time.Now().After(row.ExpirationTime) {
		return fmt.Errorf("验证码已过期")
	}
	return nil
}

package logic

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"knowsource/api/internal/svc"
	"knowsource/api/internal/types"
	"knowsource/api/internal/utils"
	"knowsource/common/response"
	"knowsource/model"

	"github.com/zeromicro/go-zero/core/logx"
	"golang.org/x/exp/rand"

	"gopkg.in/gomail.v2"
)

type SendemailcodeLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewSendemailcodeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SendemailcodeLogic {
	return &SendemailcodeLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *SendemailcodeLogic) Sendemailcode(req *types.SendemailcodeRequest, clientIP string) (resp *types.Response, err error) {
	// Generate a random verification code

	if req.Type == "register" || req.Type == "forget" {
		if !utils.VerifyCaptcha(req.CaptchaId, req.Captcha) {
			logx.Infof("验证码错误: %s, %s", req.CaptchaId, req.Captcha)
			return &types.Response{
				Code:    response.ParameterErrorCode,
				Message: "验证码错误",
				Info:    req.Captcha,
			}, nil
		}
	}

	//set redis ip key to block DDOS

	redisKey := fmt.Sprintf("email_code:%s", clientIP)
	k, err := l.svcCtx.RedisClient.Get(redisKey)
	if err != nil {
		logx.Errorw("sendemailcode redis get error", logx.Field("err", err))
		// return &types.Response{
		// 	Code:    response.ServerErrorCode,
		// 	Message: "发送邮件，连接失败",
		// 	Info:    err.Error(),
		// }, nil
	} else {
		if k != "" {
			return &types.Response{
				Code:    response.ParameterErrorCode,
				Message: "验证码发送太频繁,请1分钟后再试",
				Info:    "请稍后再试(R)",
			}, nil
		}
	}
	l.svcCtx.RedisClient.Setex(redisKey, "1", 60)

	//避免频繁发送验证码，安全

	verificationCode, err := l.svcCtx.VerificationCodesModel.FindOneByEmail(l.ctx, req.Email)
	if err == nil {
		if verificationCode.ExpirationTime.Add(-time.Minute * 10).After(time.Now().Add(-time.Minute * 1)) {
			return &types.Response{
				Code:    response.ParameterErrorCode,
				Message: "验证码发送太频繁,请1分钟后再试",
				Info:    "请稍后再试",
			}, nil
		}
	}

	code := generateVerificationCode()

	// Send the verification code to the provided email address
	err = sendVerificationEmail(l.svcCtx, req.Email, req.Type, code)
	if err != nil {
		return &types.Response{
			Code:    response.ServerErrorCode,
			Message: "验证码发送失败",
			Info:    err.Error(),
		}, nil
	}

	// Store the verification code and email in a database or cache
	err = storeVerificationData(l.svcCtx, req.Email, code, clientIP)
	if err != nil {
		logx.Error("验证码存储失败: %s", err.Error())
		return nil, err
	}

	resp = &types.Response{
		Code:    response.SuccessCode,
		Message: "验证码发送成功",
	}

	return resp, nil
}

func generateVerificationCode() string {
	// Generate a random 6-digit verification code
	seed := uint64(time.Now().UnixNano())
	rand.Seed(seed)
	code := rand.Intn(900000) + 100000
	return strconv.Itoa(code)
}

func SendStopInstanceEmail(svcCtx *svc.ServiceContext, email, instanceName string) error {
	logx.Infof("sendStopInstanceEmail start===========")

	mailAccount := svcCtx.Config.Mail.MailAccount
	mailPassword := svcCtx.Config.Mail.MailPass
	mailHost := svcCtx.Config.Mail.MailHost
	mailPort := svcCtx.Config.Mail.MailPort

	title := svcCtx.EmailConfig.StopInstanceMailTitle
	content := svcCtx.EmailConfig.StopInstanceMailContent

	if title == "" {
		title = "Out of balance"
	}
	if content == "" {
		content = "<strong>Your balance is out of balance, please recharge.</strong><br>Instance Name: %s is stopped"
	}

	m := gomail.NewMessage()
	m.SetHeader("From", mailAccount)
	m.SetHeader("To", email)
	m.SetHeader("Subject", title)
	m.SetBody("text/html", fmt.Sprintf(content, instanceName))

	d := gomail.NewDialer(mailHost, mailPort, mailAccount, mailPassword)
	// 163 SMTP 走 465 时通常需要 SSL；也可通过配置显式开启
	if svcCtx.Config.Mail.MailSSL || mailPort == 465 {
		d.SSL = true
	}

	if err := d.DialAndSend(m); err != nil {
		logx.Infof("sendStopInstanceEmail failed to send verification email: %v", err)
		return fmt.Errorf("failed to send verification email: %v", err)
	}
	logx.Infof("sendStopInstanceEmail success")

	return nil
}

func sendVerificationEmail(svcCtx *svc.ServiceContext, email, emailtype, code string) error {
	logx.Infof("sendVerificationEmail start===========")

	mailAccount := svcCtx.Config.Mail.MailAccount
	mailPassword := svcCtx.Config.Mail.MailPass
	mailHost := svcCtx.Config.Mail.MailHost
	mailPort := svcCtx.Config.Mail.MailPort

	title := ""
	content := ""
	if emailtype == "register" {
		title = svcCtx.EmailConfig.VerifyMailTitle
		content = svcCtx.EmailConfig.VerifyMailContent

		if title == "" {
			title = "Verification Email Address"
		}
		if content == "" {
			content = "<strong>Please click the verification link <a href=%s/home/email?email=%s&code=%s>%s/home/email?email=%s&code=%s</a> to verify your email address.</strong>"
		}
	} else if emailtype == "forget" {
		title = svcCtx.EmailConfig.ForgetPasswordMailTitle
		content = svcCtx.EmailConfig.ForgetPasswordMailContent

		if title == "" {
			title = "Email Verification Code"
		}
		if content == "" {
			content = "<strong>Plases login with email code : %s</strong>"
		}
	} else {
		title = svcCtx.EmailConfig.LoginMailTitle
		content = svcCtx.EmailConfig.LoginMailContent

		if title == "" {
			title = "Email Verification Code"
		}
		if content == "" {
			content = "<strong>Your verification code is: %s</strong>"
		}
	}

	m := gomail.NewMessage()
	m.SetHeader("From", mailAccount)
	m.SetHeader("To", email)
	m.SetHeader("Subject", title)
	if emailtype == "register" {
		m.SetBody("text/html", fmt.Sprintf(content, svcCtx.Config.Fca.Url, email, code, svcCtx.Config.Fca.Url, email, code))
	} else if emailtype == "forget" {
		m.SetBody("text/html", fmt.Sprintf(content, code))
	} else {
		m.SetBody("text/html", fmt.Sprintf(content, code))
	}

	d := gomail.NewDialer(mailHost, mailPort, mailAccount, mailPassword)
	// 163 SMTP 走 465 时通常需要 SSL；也可通过配置显式开启
	if svcCtx.Config.Mail.MailSSL || mailPort == 465 {
		d.SSL = true
	}

	if err := d.DialAndSend(m); err != nil {
		return fmt.Errorf("failed to send verification email: %v", err)
	}
	logx.Infof("send mail success")

	return nil
}

func storeVerificationData(svcCtx *svc.ServiceContext, email, code, clientIP string) error {
	// Create a new VerificationCodes struct
	verificationCode := model.VerificationCodes{
		TargetType:     "email",
		TargetValue:    email,
		Code:           code,
		Ip:             clientIP,
		ExpirationTime: time.Now().Add(10 * time.Minute),
	}

	// Insert the verification code into the database
	_, err := svcCtx.VerificationCodesModel.Insert(context.Background(), &verificationCode)
	if err != nil {
		return fmt.Errorf("failed to store verification code: %v", err)
	}

	return nil
}

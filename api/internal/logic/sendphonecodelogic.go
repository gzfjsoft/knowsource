package logic

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"strings"
	"time"

	"knowsource/api/internal/svc"
	"knowsource/api/internal/types"
	"knowsource/common/response"
	"knowsource/model"

	"github.com/volcengine/volc-sdk-golang/service/sms"
	"github.com/zeromicro/go-zero/core/logx"

	openapi "github.com/alibabacloud-go/darabonba-openapi/v2/client"
	dysmsapi "github.com/alibabacloud-go/dysmsapi-20170525/v4/client"
	"github.com/alibabacloud-go/tea/tea"
)

type SendphonecodeLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewSendphonecodeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SendphonecodeLogic {
	return &SendphonecodeLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *SendphonecodeLogic) Sendphonecode(req *types.SendphonecodeRequest, ip string) (resp *types.Response, err error) {

	redisKey := fmt.Sprintf("sendphonecode:%s", ip)
	k, err := l.svcCtx.RedisClient.Get(redisKey)
	if err != nil {
		logx.Errorw("sendphonecode redis get error", logx.Field("err", err))
		//return RspNew(response.ServerErrorCode, "发送邮件，连接失败", err.Error()), nil
	} else {
		if k != "" {
			return &types.Response{
				Code:    response.ParameterErrorCode,
				Message: "验证码发送太频繁,请30秒后再试",
				Info:    "请稍后再试(R)",
			}, nil
		}
	}
	l.svcCtx.RedisClient.Setex(redisKey, "1", 30)

	///==================

	// Generate a random 6-digit code
	code := fmt.Sprintf("%06d", rand.Intn(1000000))
	l.Logger.Infof("Generated verification code: %s for phone number: %s", code, req.Phonenum)

	// Save the code to the database
	verificationCode := &model.VerificationCodes{
		TargetType:     "phone",
		TargetValue:    req.Phonenum,
		Code:           code,
		Ip:             ip,
		ExpirationTime: time.Now().Add(10 * time.Minute),
	}
	_, err = l.svcCtx.VerificationCodesModel.Insert(l.ctx, verificationCode)
	if err != nil {
		l.Logger.Errorf("Failed to save verification code: %v", err)
		return &types.Response{
			Code:    response.ServerErrorCode,
			Message: "Failed to save verification code",
		}, err
	}
	l.Logger.Info("Verification code saved to database")

	// Choose SMS service based on configuration

	var smsErr error

	if l.svcCtx.Config.SMS.Supplier == "baishan" {
		_, smsErr = l.sendSMSBaishan(req.Phonenum, code)
	} else if l.svcCtx.Config.SMS.Supplier == "volc" {
		smsErr = l.sendSMSVolc(req.Phonenum, code)
	} else {
		smsErr = l.sendSMSAlibaba(req.Phonenum, code)
	}

	if smsErr != nil {
		l.Logger.Errorf("Failed to send SMS: %v", smsErr)
		return &types.Response{
			Code:    response.ServerErrorCode,
			Message: "Failed to send SMS",
			Info:    fmt.Sprintf("SMS send error: %s", smsErr.Error()),
		}, smsErr
	}

	l.Logger.Infof("SMS sent successfully to %s", req.Phonenum)
	return &types.Response{
		Code:    response.SuccessCode,
		Message: "Verification code sent successfully",
	}, nil
}

func (l *SendphonecodeLogic) sendSMSVolc(phoneNumber, code string) error {
	cfg := l.svcCtx.Config.SMS.Volc
	if cfg.AccessKey == "" || cfg.SecretKey == "" || cfg.SmsAccount == "" || cfg.TemplateID == "" {
		return fmt.Errorf("未完整配置 SMS.Volc (AccessKey, SecretKey, SmsAccount, TemplateID)")
	}
	sign := cfg.Sign
	if sign == "" {
		sign = "短信服务"
	}
	sms.DefaultInstance.Client.SetAccessKey(cfg.AccessKey)
	sms.DefaultInstance.Client.SetSecretKey(cfg.SecretKey)
	reqsms := &sms.SmsRequest{
		SmsAccount:    cfg.SmsAccount,
		Sign:          sign,
		TemplateID:    cfg.TemplateID,
		TemplateParam: fmt.Sprintf(`{"code":"%s"}`, code),
		PhoneNumbers:  phoneNumber,
		Tag:           "tag",
	}
	smsResponse, _, err := sms.DefaultInstance.Send(reqsms)
	if err != nil {
		return err
	}

	if smsResponse.ResponseMetadata.Error != nil {
		return fmt.Errorf("SMS service returned an error: %s", smsResponse.ResponseMetadata.Error.Message)
	}

	return nil
}

func (l *SendphonecodeLogic) sendSMSAlibaba(phoneNumber, code string) error {
	accessKeyId := l.svcCtx.Config.Aliyun.AccessKeyId
	accessKeySecret := l.svcCtx.Config.Aliyun.AccessKeySecret
	TemplateCode := l.svcCtx.Config.SMS.LoginTemplateCode
	CompanyName := l.svcCtx.Config.SMS.CompanyName
	if accessKeyId == "" || accessKeySecret == "" || TemplateCode == "" || CompanyName == "" {
		return fmt.Errorf("未完整配置 Aliyun AccessKey 或 SMS.LoginTemplateCode / SMS.CompanyName")
	}

	config := &openapi.Config{
		AccessKeyId:     &accessKeyId,
		AccessKeySecret: &accessKeySecret,
		Endpoint:        tea.String("dysmsapi.aliyuncs.com"),
	}

	client, err := dysmsapi.NewClient(config)
	if err != nil {
		return err
	}

	sendSmsRequest := &dysmsapi.SendSmsRequest{
		PhoneNumbers:  tea.String(phoneNumber),
		SignName:      tea.String(CompanyName),
		TemplateCode:  tea.String(TemplateCode),
		TemplateParam: tea.String(fmt.Sprintf("{\"code\":\"%s\"}", code)),
	}

	sendresponse, err := client.SendSms(sendSmsRequest)
	if err != nil {
		return err
	}

	if *sendresponse.Body.Code != "OK" {
		return fmt.Errorf("SMS send failed with code: %s, message: %s", *sendresponse.Body.Code, *sendresponse.Body.Message)
	}

	//
	// 	"statusCode": 200,
	//    "body": {
	//       "Code": "isv.AMOUNT_NOT_ENOUGH",
	//       "Message": "账户余额不足",
	//       "RequestId": "2E3C074F-3C9F-5664-B09F-3086CC8769B7"
	//    }

	// "statusCode": 200,
	// "body": {
	//    "BizId": "645924827285632285^0",
	//    "Code": "OK",
	//    "Message": "OK",
	//    "RequestId": "6BB40AEE-D820-5F61-88FA-2607D7C1EB6A"
	// }

	l.Logger.Infof("resp", sendresponse)

	return nil
}

// SendSMS 发送短信的函数
func (l *SendphonecodeLogic) sendSMSBaishan(phone, param string) (string, error) {
	token := strings.TrimSpace(l.svcCtx.Config.SMS.Baishan.Token)
	template := strings.TrimSpace(l.svcCtx.Config.SMS.Baishan.Template)
	if token == "" || template == "" {
		return "", fmt.Errorf("未完整配置 SMS.Baishan (Token, Template)")
	}
	// 构建请求体数据
	reqData := struct {
		Token  string `json:"token"`
		Params struct {
			Phone    string `json:"phone"`
			Template string `json:"template"`
			Param    string `json:"param"`
		} `json:"params"`
	}{
		Token: token,
		Params: struct {
			Phone    string `json:"phone"`
			Template string `json:"template"`
			Param    string `json:"param"`
		}{
			Phone:    phone,
			Template: template,
			Param:    param,
		},
	}
	// 将请求体数据编码为JSON格式
	reqBody, err := json.Marshal(reqData)
	if err != nil {
		return "", err
	}
	// 创建HTTP请求
	req, err := http.NewRequest("POST", "https://msgg.bs58i.baishancdnx.com/api/app/1.0/msgg/submitsms", bytes.NewBuffer(reqBody))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	// 发送HTTP请求
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	// 读取响应体
	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return string(respBody), nil
}

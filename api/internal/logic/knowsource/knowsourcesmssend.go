package knowsource

import (
	"fmt"
	"strings"

	"knowsource/api/internal/svc"

	openapi "github.com/alibabacloud-go/darabonba-openapi/v2/client"
	dysmsapi "github.com/alibabacloud-go/dysmsapi-20170525/v4/client"
	"github.com/alibabacloud-go/tea/tea"
)

// KnowsourceSendSMSCode 使用阿里云短信发送验证码（与 SendphonecodeLogic 一致，供找回密码等复用）
func KnowsourceSendSMSCode(svcCtx *svc.ServiceContext, phoneNumber, code string) error {
	phoneNumber = strings.TrimSpace(phoneNumber)
	if phoneNumber == "" {
		return fmt.Errorf("手机号为空")
	}
	sup := strings.TrimSpace(svcCtx.Config.SMS.Supplier)
	if sup != "" && sup != "alibaba" {
		return fmt.Errorf("当前短信供应商为 %s， 找回密码仅实现 alibaba 通道，请在配置中将 SMS.Supplier 设为 alibaba 或留空", sup)
	}

	accessKeyId := svcCtx.Config.Aliyun.AccessKeyId
	accessKeySecret := svcCtx.Config.Aliyun.AccessKeySecret
	templateCode := svcCtx.Config.SMS.LoginTemplateCode
	companyName := svcCtx.Config.SMS.CompanyName
	if accessKeyId == "" || accessKeySecret == "" || templateCode == "" || companyName == "" {
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
		SignName:      tea.String(companyName),
		TemplateCode:  tea.String(templateCode),
		TemplateParam: tea.String(fmt.Sprintf(`{"code":"%s"}`, code)),
	}
	sendresponse, err := client.SendSms(sendSmsRequest)
	if err != nil {
		return err
	}
	if sendresponse.Body == nil || sendresponse.Body.Code == nil || *sendresponse.Body.Code != "OK" {
		msg := ""
		if sendresponse.Body != nil && sendresponse.Body.Message != nil {
			msg = *sendresponse.Body.Message
		}
		return fmt.Errorf("短信发送失败: %s", msg)
	}
	return nil
}

package logic

import (
	"context"

	"knowsource/api/internal/svc"
	"knowsource/api/internal/types"
	"knowsource/api/internal/utils"
	"knowsource/common/response"

	"github.com/zeromicro/go-zero/core/logx"
)

type CaptchaLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewCaptchaLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CaptchaLogic {
	return &CaptchaLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CaptchaLogic) Captcha() (resp *types.CaptchaResponse, err error) {

	var aresp types.CaptchaResponse
	// Create a driver for the captcha
	// driver := base64Captcha.NewDriverString(
	// 	80,  // height
	// 	240, // width
	// 	0,   // no noise
	// 	base64Captcha.OptionShowHollowLine,
	// 	5,                                     // length
	// 	"234567890abcdefghijklmnopqrstuvwxyz", // source
	// 	&color.RGBA{R: 0, G: 0, B: 0, A: 0},   // foreground color
	// 	nil,                                   // fonts
	// 	[]string{},
	// )

	// Create a captcha using the driver
	//captcha := base64Captcha.NewCaptcha(driver, base64Captcha.DefaultMemStore)
	//GenerateCaptcha()

	// Generate the captcha
	//id, b64s, _, err := captcha.Generate()
	id, b64s, err := utils.GenerateCaptcha()
	if err != nil {
		aresp.Code = response.ServerErrorCode
		aresp.Message = "Failed to generate captcha"
		aresp.Info = err.Error()
		return &aresp, nil
	}

	// Return the captcha response
	aresp.Code = response.SuccessCode
	aresp.Message = "Success"
	aresp.Info = "Captcha generated successfully"
	aresp.Data.CaptchaId = id
	aresp.Data.ImageBase64 = b64s
	return &aresp, nil

}

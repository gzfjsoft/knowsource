package logic

import (
	"context"
	"fmt"
	"knowsource/common/response"
	"regexp"

	"knowsource/api/internal/logic/common"
	"knowsource/api/internal/svc"
	"knowsource/api/internal/types"
	"knowsource/model"

	"github.com/zeromicro/go-zero/core/logx"
)

type RegisterLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewRegisterLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RegisterLogic {
	return &RegisterLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *RegisterLogic) Register(req *types.RegisterRequest) (resp response.Response) {

	if req.Email != "" {
		emailRegex := regexp.MustCompile(`^[a-zA-Z0-9.!#$%&'*+/=?^_` + "`" + `{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$`)
		if !emailRegex.MatchString(req.Email) {
			return response.Fail(response.InvalidRequestParamCode, "邮箱格式不正确")
		}
		_, err := l.svcCtx.UsersModel.FindOneByEmail(l.ctx, req.Email)
		if err != nil && err != model.ErrNotFound {
			return response.Error(err.Error())
		}
		if err != model.ErrNotFound {
			return response.Fail(response.InvalidRequestParamCode, "你的填写的邮箱已注册")
		}

	}
	_, err := l.svcCtx.UsersModel.FindOneByUsername(l.ctx, req.Username)
	if err != nil && err != model.ErrNotFound {
		return response.Error(err.Error())
	}
	if err != model.ErrNotFound {
		return response.Fail(response.InvalidRequestParamCode, "你的填写的用户名已注册")
	}

	if req.Phone != "" {
		_, err = l.svcCtx.UsersModel.FindOneByPhone(l.ctx, req.Phone)
		if err != nil && err != model.ErrNotFound {
			return response.Error(err.Error())
		}

		if err != model.ErrNotFound {
			return response.Fail(response.InvalidRequestParamCode, "你的填写的手机号已注册")
		}
	}

	if req.Nickname == "" {
		req.Nickname = req.Username
	}

	var userId uint64
	var respData *types.RegisterResponseData

	//注册一个用户

	userId, err = l.svcCtx.UsersModel.RegisterNewUserinDB(l.ctx,
		req.Username,
		req.Nickname,
		req.Email,
		req.Phone,
		req.Password,
		"",
		"",
		"",
		"",
		"",
		l.svcCtx.Config.Salt,
	)

	if err != nil {
		return response.ErrorWithInfo("RegisterNewUserinDB", err.Error())
	}

	// 	userRolesModel := l.svcCtx.UserRolesModel.WithSession(session)
	// 	_, err = userRolesModel.Insert(l.ctx, &model.UserRoles{
	// 		UserId:     uint64(userId),
	// 		RoleId:     1,
	// 		AssignedBy: 1,
	// 		AssignedAt: time.Now(),
	// 	})

	// 	if err != nil {
	// 		return err
	// 	}

	// 	return nil
	// })

	// if err != nil {
	// 	return response.Error(err.Error())
	// }

	// Create initial balance with transaction

	//赠送10块
	fee := 1000
	balanceLogic := common.NewUserBalancesLogic(l.ctx, l.svcCtx)
	err, _ = balanceLogic.ManualAdjust(uint64(userId), 1, int64(fee), "CNY", "注册赠送", fmt.Sprintf("注册用户，赠送价值%d元体验额度", fee/100))
	if err != nil {
		logx.Error("ManualAdjust in registerlogic", err.Error())
	}

	respData = &types.RegisterResponseData{
		UserId:   uint64(userId),
		Username: req.Username,
		Email:    req.Email,
		Phone:    req.Phone,
		Nickname: req.Nickname,
	}

	logx.Info(respData, req)

	// Get organization info

	orgs, err := l.svcCtx.OrgsUsersModel.FindAllByUserId(l.ctx, uint64(userId))
	if err != nil {
		return response.Error(err.Error())
	}

	var orgUser []types.Org
	for _, aorg := range *orgs {
		orgUser = append(orgUser, types.Org{
			OrgId:   aorg.OrgId,
			OrgName: aorg.OrgName,
			Role:    aorg.Role,
		})
	}

	code := generateVerificationCode()
	err = sendVerificationEmail(l.svcCtx, req.Email, "register", code)
	if err != nil {
		logx.Infof("注册验证码发送失败: %s", err.Error())
	}
	if err == nil {
		err = storeVerificationData(l.svcCtx, req.Email, code, "")
		if err != nil {
			logx.Error("登陆验证码存储失败: %s", err.Error())
		}
	}

	respData.Orgs = orgUser
	logx.Info(respData, req)

	return response.OK(respData)
}

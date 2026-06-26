package knowsource

import (
	"context"
	"strings"

	"knowsource/api/internal/svc"
	"knowsource/api/internal/types"
	"knowsource/common/cryptx"
	"knowsource/common/response"
	"knowsource/model"

	"github.com/zeromicro/go-zero/core/logx"
)

type KnowsourcePasswordResetConfirmLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewKnowsourcePasswordResetConfirmLogic(ctx context.Context, svcCtx *svc.ServiceContext) *KnowsourcePasswordResetConfirmLogic {
	return &KnowsourcePasswordResetConfirmLogic{Logger: logx.WithContext(ctx), ctx: ctx, svcCtx: svcCtx}
}

func (l *KnowsourcePasswordResetConfirmLogic) KnowsourcePasswordResetConfirm(req *types.KnowsourcePasswordResetConfirmRequest) (resp *types.Response, err error) {
	if req == nil {
		return &types.Response{Code: response.ParameterErrorCode, Message: "参数不能为空"}, nil
	}
	clientId := strings.TrimSpace(req.ClientId)
	ch := strings.ToLower(strings.TrimSpace(req.Channel))
	newPwd := req.NewPassword
	code := strings.TrimSpace(req.Code)
	if clientId == "" || (ch != "email" && ch != "phone") || newPwd == "" || code == "" {
		return &types.Response{Code: response.ParameterErrorCode, Message: "参数不完整"}, nil
	}
	if len(newPwd) < 6 {
		return &types.Response{Code: response.ParameterErrorCode, Message: "新密码至少 6 位"}, nil
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
	if e := KnowsourceVerifyStoredCode(l.ctx, l.svcCtx, targetKnowsourcePwdReset, resetKey, code); e != nil {
		return &types.Response{Code: response.ParameterErrorCode, Message: e.Error()}, nil
	}

	hash := cryptx.PasswordEncrypt(l.svcCtx.Config.Salt, newPwd)
	pwdRow, e := l.svcCtx.EmpPasswordModel.FindOneByClientIdEmpCode(l.ctx, clientId, emp.FempCode)
	if e != nil && e == model.ErrNotFound {
		_, e = l.svcCtx.EmpPasswordModel.Insert(l.ctx, &model.EmpPassword{
			ClientId: clientId,
			EmpCode:  emp.FempCode,
			Password: hash,
		})
	} else if e == nil && pwdRow != nil {
		pwdRow.Password = hash
		e = l.svcCtx.EmpPasswordModel.Update(l.ctx, pwdRow)
	}
	if e != nil {
		return &types.Response{Code: response.ServerErrorCode, Message: "更新密码失败", Info: e.Error()}, nil
	}
	return &types.Response{Code: response.SuccessCode, Message: "密码已重置"}, nil
}

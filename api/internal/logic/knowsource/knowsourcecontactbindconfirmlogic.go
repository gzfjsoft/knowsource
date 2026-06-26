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

type KnowsourceContactBindConfirmLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewKnowsourceContactBindConfirmLogic(ctx context.Context, svcCtx *svc.ServiceContext) *KnowsourceContactBindConfirmLogic {
	return &KnowsourceContactBindConfirmLogic{Logger: logx.WithContext(ctx), ctx: ctx, svcCtx: svcCtx}
}

func (l *KnowsourceContactBindConfirmLogic) KnowsourceContactBindConfirm(req *types.KnowsourceContactBindConfirmRequest) (resp *types.Response, err error) {
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
	code := strings.TrimSpace(req.Code)
	if (ch != "email" && ch != "phone") || code == "" {
		return &types.Response{Code: response.ParameterErrorCode, Message: "channel 或 code 无效"}, nil
	}

	emp, e := l.svcCtx.FrEmpModel.FindOneByClientIdFempCode(l.ctx, clientId, empCode)
	if e != nil {
		if e == model.ErrNotFound {
			return &types.Response{Code: response.UserNotExistCode, Message: "员工不存在"}, nil
		}
		return &types.Response{Code: response.ServerErrorCode, Message: "Database error", Info: e.Error()}, nil
	}

	if ch == "email" {
		em := strings.TrimSpace(strings.ToLower(req.NewEmail))
		if em == "" {
			return &types.Response{Code: response.ParameterErrorCode, Message: "请填写 newEmail"}, nil
		}
		tv := clientId + "|" + empCode + "|" + em
		if e := KnowsourceVerifyStoredCode(l.ctx, l.svcCtx, targetKnowsourceBindEmail, tv, code); e != nil {
			return &types.Response{Code: response.ParameterErrorCode, Message: e.Error()}, nil
		}
		if o, e := l.svcCtx.FrEmpModel.FindOneByClientIdEmail(l.ctx, clientId, em); e == nil && o != nil && o.FempCode != empCode {
			return &types.Response{Code: response.ConflictCode, Message: "该邮箱已被其他员工使用"}, nil
		} else if e != nil && e != model.ErrNotFound {
			return &types.Response{Code: response.ServerErrorCode, Message: "Database error", Info: e.Error()}, nil
		}
		emp.Email = em
	} else {
		ph := strings.TrimSpace(req.NewPhone)
		if ph == "" {
			return &types.Response{Code: response.ParameterErrorCode, Message: "请填写 newPhone"}, nil
		}
		tv := clientId + "|" + empCode + "|" + ph
		if e := KnowsourceVerifyStoredCode(l.ctx, l.svcCtx, targetKnowsourceBindPhone, tv, code); e != nil {
			return &types.Response{Code: response.ParameterErrorCode, Message: e.Error()}, nil
		}
		if o, e := l.svcCtx.FrEmpModel.FindOneByClientIdMobile(l.ctx, clientId, ph); e == nil && o != nil && o.FempCode != empCode {
			return &types.Response{Code: response.ConflictCode, Message: "该手机号已被其他员工使用"}, nil
		} else if e != nil && e != model.ErrNotFound {
			return &types.Response{Code: response.ServerErrorCode, Message: "Database error", Info: e.Error()}, nil
		}
		emp.Mobile = ph
	}

	if e := l.svcCtx.FrEmpModel.Update(l.ctx, emp); e != nil {
		return &types.Response{Code: response.ServerErrorCode, Message: "更新失败", Info: e.Error()}, nil
	}
	return &types.Response{Code: response.SuccessCode, Message: "绑定成功"}, nil
}

package knowsource

import (
	"context"
	"database/sql"
	"strings"
	"time"

	"knowsource/api/internal/svc"
	"knowsource/api/internal/types"
	"knowsource/common/response"
	"knowsource/model"

	"github.com/zeromicro/go-zero/core/logx"
)

type KnowsourceTenantVerifyEmailLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewKnowsourceTenantVerifyEmailLogic(ctx context.Context, svcCtx *svc.ServiceContext) *KnowsourceTenantVerifyEmailLogic {
	return &KnowsourceTenantVerifyEmailLogic{Logger: logx.WithContext(ctx), ctx: ctx, svcCtx: svcCtx}
}

func (l *KnowsourceTenantVerifyEmailLogic) KnowsourceTenantVerifyEmail(req *types.KnowsourceTenantVerifyEmailRequest) (resp *types.Response, err error) {
	if req == nil {
		return &types.Response{Code: response.ParameterErrorCode, Message: "参数不能为空"}, nil
	}
	clientId := strings.TrimSpace(req.ClientId)
	code := strings.TrimSpace(req.Code)
	if clientId == "" || code == "" {
		return &types.Response{Code: response.ParameterErrorCode, Message: "clientId、code 不能为空"}, nil
	}

	// 查找租户信息
	row, e := l.svcCtx.ClientModel.FindOneByClientId(l.ctx, clientId)
	if e != nil || row == nil {
		if e == model.ErrNotFound {
			return &types.Response{Code: response.NotFoundCode, Message: "租户不存在"}, nil
		}
		return &types.Response{Code: response.ServerErrorCode, Message: "Database error", Info: e.Error()}, nil
	}

	if e := KnowsourceVerifyStoredCode(l.ctx, l.svcCtx, targetKnowsourceTenantVerify, clientId, code); e != nil {
		return &types.Response{Code: response.ParameterErrorCode, Message: e.Error()}, nil
	}
	if row.Status != 0 {
		return &types.Response{Code: response.SuccessCode, Message: "该租户已验证或无需重复验证"}, nil
	}

	row.Status = 1
	row.VerifiedAt = sql.NullTime{Time: time.Now(), Valid: true}
	if e := l.svcCtx.ClientModel.Update(l.ctx, row); e != nil {
		return &types.Response{Code: response.ServerErrorCode, Message: "更新失败", Info: e.Error()}, nil
	}
	return &types.Response{Code: response.SuccessCode, Message: "验证成功"}, nil
}

package admin

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"

	"knowsource/api/internal/svc"
	"knowsource/api/internal/types"
	"knowsource/common/response"
	"knowsource/consts"

	"github.com/zeromicro/go-zero/core/logx"
)

type AdminAduitIdcardLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewAdminAduitIdcardLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AdminAduitIdcardLogic {
	return &AdminAduitIdcardLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *AdminAduitIdcardLogic) AdminAduitIdcard(req *types.AdminAduitIdcardRequest) (resp *types.Response, err error) {
	role, _ := l.ctx.Value("role").(string)
	uid, _ := l.ctx.Value("uid").(json.Number).Int64()

	if role != consts.SUPER_ADMIN {
		return &types.Response{
			Code:    response.ServerErrorCode,
			Message: "只有超级管理员可以审核身份证",
		}, nil
	}

	userAuthIdcard, err := l.svcCtx.UserAuthIdcardModel.FindOne(l.ctx, req.Id)
	if err != nil {
		return &types.Response{
			Code:    response.ServerErrorCode,
			Message: "身份证审核记录不存在",
		}, nil
	}

	if userAuthIdcard.AuditStatus != 1 {
		return &types.Response{
			Code:    response.ServerErrorCode,
			Message: "身份证审核记录不在审核中状态",
		}, nil
	}

	userAuthIdcard.AuditStatus = req.AuditStatus
	userAuthIdcard.FailReason = sql.NullString{String: req.FailReason, Valid: true}
	userAuthIdcard.AuditUserId = uint64(uid)
	userAuthIdcard.AuditAt = time.Now()

	err = l.svcCtx.UserAuthIdcardModel.Update(l.ctx, userAuthIdcard)

	return &types.Response{
		Code:    response.SuccessCode,
		Message: "身份证审核成功",
	}, nil
}

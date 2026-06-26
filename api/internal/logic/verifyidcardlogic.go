package logic

import (
	"context"
	"encoding/json"
	"time"

	"knowsource/api/internal/svc"
	"knowsource/api/internal/types"
	"knowsource/common/response"
	"knowsource/model"

	"github.com/zeromicro/go-zero/core/logx"
)

type VerifyIdcardLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewVerifyIdcardLogic(ctx context.Context, svcCtx *svc.ServiceContext) *VerifyIdcardLogic {
	return &VerifyIdcardLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

/*
UserAuthIdcard struct {
Id               uint64         `db:"id"`

		UserId           int64          `db:"user_id"`
		ReqName          string         `db:"req_name"`       // 用户输入姓名
		ReqIdcardNum     string         `db:"req_idcard_num"` // 用户输入身份证号
		ImageFront       string         `db:"image_front"`
		ImageBack        string         `db:"image_back"`
		AuditStatus      int64          `db:"audit_status"` // 1:审核中，2：审核通过，3：审核未通过，0：未上传审核
		FailReason       sql.NullString `db:"fail_reason"`
		State            int64          `db:"state"`
		CreatedAt        time.Time      `db:"created_at"`
		UpdatedAt        time.Time      `db:"updated_at"`
		Address          string         `db:"Address"`
		Birthday         string         `db:"birthday"`
		ValidityPeriod   string         `db:"validity_period"`
		Name             string         `db:"name"`              // 身份证姓名
		Gender           int64          `db:"gender"`            // 身份证性别
		IssuingAuthority string         `db:"issuing_authority"` // 发证机关
		Ethnicity        string         `db:"ethnicity"`         // 民族
		IdcardNum        string         `db:"idcard_num"`        // 身份证号
		AuditUserId      uint64         `db:"audit_user_id"`     // 审核人ID
		AuditAt          time.Time      `db:"audit_at"`          // 审核时间
	}
*/
func (l *VerifyIdcardLogic) VerifyIdcard(req *types.VerifyRequest) (resp *types.Response, err error) {

	uid, _ := l.ctx.Value("uid").(json.Number).Int64()

	_, err = l.svcCtx.UserAuthIdcardModel.Insert(l.ctx, &model.UserAuthIdcard{
		UserId:       int64(uid),
		ReqName:      req.Name,
		ReqIdcardNum: req.Idcard,
		ImageFront:   req.IdcardFront,
		ImageBack:    req.IdcardBack,
		AuditStatus:  1,
		State:        1,
		AuditAt:      time.Now(),
	})

	if err != nil {
		return &types.Response{
			Message: "insert user auth idcard failed",
			Code:    response.ServerErrorCode,
		}, nil
	}

	return &types.Response{
		Message: "insert user auth idcard success",
		Code:    response.SuccessCode,
	}, nil
}

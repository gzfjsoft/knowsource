package admin

import (
	"context"

	"knowsource/api/internal/svc"
	"knowsource/api/internal/types"
	"knowsource/common/response"
	"knowsource/consts"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetVerifyIdcardListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetVerifyIdcardListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetVerifyIdcardListLogic {
	return &GetVerifyIdcardListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetVerifyIdcardListLogic) GetVerifyIdcardList(req *types.GetVerifyIdcardListRequest) (resp *types.GetVerifyIdcardListResponse, err error) {

	role, _ := l.ctx.Value("role").(string)
	if role != consts.SUPER_ADMIN {
		return &types.GetVerifyIdcardListResponse{
			Response: types.Response{Message: "permission denied", Code: response.ServerErrorCode},
		}, nil
	}

	userAuthIdcardList, err := l.svcCtx.UserAuthIdcardModel.GetVerifyIdcardList(l.ctx, req.Page, req.PageSize)
	if err != nil {
		return &types.GetVerifyIdcardListResponse{
			Response: types.Response{Message: "get verify idcard list failed", Code: response.ServerErrorCode},
		}, nil
	}
	// UserAuthIdcard struct {
	// 		Id               uint64         `db:"id"`
	// 		UserId           int64          `db:"user_id"`
	// 		ReqName          string         `db:"req_name"`       // 用户输入姓名
	// 		ReqIdcardNum     string         `db:"req_idcard_num"` // 用户输入身份证号
	// 		ImageFront       string         `db:"image_front"`
	// 		ImageBack        string         `db:"image_back"`
	// 		AuditStatus      int64          `db:"audit_status"` // 1:审核中，2：审核通过，3：审核未通过，0：未上传审核
	// 		FailReason       sql.NullString `db:"fail_reason"`
	// 		State            int64          `db:"state"`
	// 		CreatedAt        time.Time      `db:"created_at"`
	// 		UpdatedAt        time.Time      `db:"updated_at"`
	// 		Address          string         `db:"Address"`
	// 		Birthday         string         `db:"birthday"`
	// 		ValidityPeriod   string         `db:"validity_period"`
	// 		Name             string         `db:"name"`              // 身份证姓名
	// 		Gender           int64          `db:"gender"`            // 身份证性别
	// 		IssuingAuthority string         `db:"issuing_authority"` // 发证机关
	// 		Ethnicity        string         `db:"ethnicity"`         // 民族
	// 		IdcardNum        string         `db:"idcard_num"`        // 身份证号
	// 		AuditUserId      uint64         `db:"audit_user_id"`     // 审核人ID
	// 		AuditAt          time.Time      `db:"audit_at"`          // 审核时间
	// 	}

	var list []types.GetVerifyIdcardListDataItem
	for _, item := range userAuthIdcardList {
		list = append(list, types.GetVerifyIdcardListDataItem{
			Id:               item.Id,
			UserId:           item.UserId,
			ReqName:          item.ReqName,
			ReqIdcardNum:     item.ReqIdcardNum,
			ImageFront:       item.ImageFront,
			ImageBack:        item.ImageBack,
			AuditStatus:      item.AuditStatus,
			FailReason:       item.FailReason.String,
			State:            item.State,
			CreatedAt:        item.CreatedAt.Unix(),
			UpdatedAt:        item.UpdatedAt.Unix(),
			Address:          item.Address,
			Birthday:         item.Birthday,
			ValidityPeriod:   item.ValidityPeriod,
			Name:             item.Name,
			Gender:           item.Gender,
			IssuingAuthority: item.IssuingAuthority,
			Ethnicity:        item.Ethnicity,
			IdcardNum:        item.IdcardNum,
			AuditUserId:      item.AuditUserId,
			AuditAt:          item.AuditAt.Unix(),
		})
	}

	return &types.GetVerifyIdcardListResponse{
		Response: types.Response{Message: "get verify idcard list success", Code: response.SuccessCode},
		Data: &types.GetVerifyIdcardListData{
			List:  list,
			Total: int64(len(list)),
		},
	}, nil
}

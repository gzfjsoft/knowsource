package organization

import (
	"context"
	"strings"

	"knowsource/api/internal/svc"
	"knowsource/api/internal/types"
	"knowsource/common/response"

	"github.com/zeromicro/go-zero/core/logx"
)

type CreateOrgLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewCreateOrgLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateOrgLogic {
	return &CreateOrgLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CreateOrgLogic) CreateOrg(req *types.CreateOrgRequest) (*types.CreateOrgResp, error) {
	if req == nil || strings.TrimSpace(req.OrgName) == "" {
		return &types.CreateOrgResp{
			Response: types.Response{
				Code:    response.InvalidRequestParamCode,
				Message: "orgName不能为空",
			},
		}, nil
	}

	// 该接口的完整实现依赖于组织表/权限模型；当前先返回可用的响应，避免 handler 引用缺失导致无法编译。
	return &types.CreateOrgResp{
		Response: types.Response{
			Code:    response.SuccessCode,
			Message: "success",
		},
		Data: types.Organization{
			OrgId:     0,
			OrgName:   strings.TrimSpace(req.OrgName),
			IsPrivate: req.IsPrivate,
			IsDefault: req.IsDefault,
		},
	}, nil
}


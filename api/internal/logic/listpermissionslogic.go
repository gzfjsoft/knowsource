package logic

import (
	"context"
	"knowsource/api/internal/svc"
	"knowsource/api/internal/types"
	"knowsource/common/response"
	"knowsource/consts"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListPermissionsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewListPermissionsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListPermissionsLogic {
	return &ListPermissionsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ListPermissionsLogic) ListPermissions(req *types.PermissionListRequest) (resp *types.PermissionListResponse) {

	sysrole, _ := l.ctx.Value("role").(string)

	if sysrole != consts.ONLY_ADMIN && sysrole != consts.SUPER_ADMIN {
		return &types.PermissionListResponse{
			Response: types.Response{
				Code:    response.UnauthorizedCode,
				Message: "没有权限",
			},
		}
	}

	resp = &types.PermissionListResponse{
		Response: types.Response{
			Code:    response.SuccessCode,
			Message: "Success",
		},
	}

	permissions, total, err := l.svcCtx.PermissionsModel.FindByFilter(l.ctx, req.Name, req.Code, req.Page, req.PageSize)
	if err != nil {
		resp.Response = types.Response{
			Code:    response.ServerErrorCode,
			Message: err.Error(),
		}
		return resp
	}

	var permList []types.Permission
	for _, perm := range permissions {
		permList = append(permList, types.Permission{
			Id:          perm.PermissionId,
			Name:        perm.PermissionName,
			Description: perm.Description.String,
			CreatedAt:   perm.CreatedAt.Unix(),
		})
	}

	resp.Response = types.Response{
		Code:    response.SuccessCode,
		Message: "Success",
	}
	resp.Data = types.PermissionListResponseData{
		Permissions: permList,
		Total:       total,
	}

	return resp
}

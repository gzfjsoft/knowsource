package logic

import (
	"context"
	"knowsource/api/internal/svc"
	"knowsource/api/internal/types"
	"knowsource/common/response"
	"knowsource/consts"
	"knowsource/model"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetPermissionLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetPermissionLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetPermissionLogic {
	return &GetPermissionLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetPermissionLogic) GetPermission(id uint64) (resp *types.PermissionResponse) {

	sysrole, _ := l.ctx.Value("role").(string)
	if sysrole != consts.ONLY_ADMIN && sysrole != consts.SUPER_ADMIN {
		return &types.PermissionResponse{
			Response: types.Response{
				Code:    response.UnauthorizedCode,
				Message: "没有权限",
			},
		}
	}

	resp = &types.PermissionResponse{
		Response: types.Response{},
	}

	permission, err := l.svcCtx.PermissionsModel.FindOne(l.ctx, id)
	if err != nil {
		if err == model.ErrNotFound {
			resp.Response = types.Response{
				Code:    response.InvalidRequestParamCode,
				Message: "Permission not found",
			}
			return resp
		}
		resp.Response = types.Response{
			Code:    response.ServerErrorCode,
			Message: err.Error(),
		}
		return resp
	}

	resp.Response = types.Response{
		Code:    response.SuccessCode,
		Message: "Success",
	}
	resp.Data = types.Permission{
		Id:          permission.PermissionId,
		Name:        permission.PermissionName,
		Description: permission.Description.String,
		CreatedAt:   permission.CreatedAt.Unix(),
	}

	return resp
}

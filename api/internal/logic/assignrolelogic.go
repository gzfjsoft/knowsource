package logic

import (
	"context"
	"encoding/json"
	"knowsource/api/internal/svc"
	"knowsource/api/internal/types"
	"knowsource/common/response"
	"knowsource/consts"
	"knowsource/model"
	"time"

	"github.com/zeromicro/go-zero/core/logx"
)

type AssignRoleLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewAssignRoleLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AssignRoleLogic {
	return &AssignRoleLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *AssignRoleLogic) AssignRole(req *types.AssignRoleRequest) (resp response.Response) {

	sysrole, _ := l.ctx.Value("role").(string)
	if sysrole != consts.ONLY_ADMIN && sysrole != consts.SUPER_ADMIN {
		return response.Fail(response.UnauthorizedCode, "没有权限")
	}

	// Check if user exists
	uid, _ := l.ctx.Value("uid").(json.Number).Int64()

	_, err := l.svcCtx.UsersModel.FindOne(l.ctx, req.UserId)
	if err != nil {
		if err == model.ErrNotFound {
			return response.Fail(response.InvalidRequestParamCode, "User not found")
		}
		return response.Error(err.Error())
	}

	// Check if role exists
	_, err = l.svcCtx.RolesModel.FindOne(l.ctx, req.RoleId)
	if err != nil {
		if err == model.ErrNotFound {
			return response.Fail(response.InvalidRequestParamCode, "Role not found")
		}
		return response.Error(err.Error())
	}

	// Check if user already has this role
	existingUserRole, err := l.svcCtx.UserRolesModel.FindOneByUserIdRoleId(l.ctx, req.UserId, req.RoleId)
	if err != nil && err != model.ErrNotFound {
		return response.Error(err.Error())
	}
	if existingUserRole != nil {
		return response.Fail(response.InvalidRequestParamCode, "User already has this role")
	}

	userRole := &model.UserRoles{
		UserId:     uint64(req.UserId),
		RoleId:     uint64(req.RoleId),
		AssignedAt: time.Now(),
		AssignedBy: uint64(uid),
	}

	_, err = l.svcCtx.UserRolesModel.Insert(l.ctx, userRole)
	if err != nil {
		return response.Error(err.Error())
	}

	return response.OK(nil)
}

package knowsource

import (
	"context"
	"strings"

	"knowsource/api/internal/svc"
	"knowsource/api/internal/types"
	"knowsource/common/response"
	"knowsource/consts"
	"knowsource/model"

	"github.com/pkg/errors"
	"github.com/zeromicro/go-zero/core/logx"
)

type DeleteFrRoleLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 删除角色
func NewDeleteFrRoleLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteFrRoleLogic {
	return &DeleteFrRoleLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *DeleteFrRoleLogic) DeleteFrRole(req *types.FrRoleDeleteRequest) (resp *types.Response, err error) {
	clientId, _ := l.ctx.Value("clientId").(string)
	clientId = strings.TrimSpace(clientId)
	if clientId == "" {
		return &types.Response{
			Code:    response.UnauthorizedCode,
			Message: "clientId不能为空，请重新登录",
		}, nil
	}

	if req.Role == "" {
		return &types.Response{
			Code:    response.ParameterErrorCode,
			Message: "角色编码不能为空",
		}, nil
	}

	// 保护：禁止删除 superadmin 角色
	if strings.TrimSpace(req.Role) == consts.SUPER_ADMIN {
		return &types.Response{
			Code:    response.ConflictCode,
			Message: "禁止删除superadmin角色",
		}, nil
	}

	// 检查角色是否存在
	_, err = l.svcCtx.FrRolesModel.FindOneByClientIdRole(l.ctx, clientId, req.Role)
	if err != nil {
		if errors.Is(err, model.ErrNotFound) {
			return &types.Response{
				Code:    response.NotFoundCode,
				Message: "角色不存在",
			}, nil
		}
		l.Logger.Errorf("查询角色失败: %v", err)
		return &types.Response{
			Code:    response.ServerErrorCode,
			Message: "Database error",
			Info:    err.Error(),
		}, nil
	}

	// 检查是否有用户关联此角色
	userCount, err := l.svcCtx.FrUserRolesModel.CountByClientIdRole(l.ctx, clientId, req.Role)
	if err != nil {
		l.Logger.Errorf("查询角色关联用户数失败: %v", err)
		return &types.Response{
			Code:    response.ServerErrorCode,
			Message: "Database error",
			Info:    err.Error(),
		}, nil
	}
	if userCount > 0 {
		return &types.Response{
			Code:    response.ParameterErrorCode,
			Message: "该角色还有关联的用户，无法删除。请先解除该角色与用户的关联关系",
		}, nil
	}

	// 删除角色
	err = l.svcCtx.FrRolesModel.DeleteByClientIdRole(l.ctx, clientId, req.Role)
	if err != nil {
		l.Logger.Errorf("删除角色失败: %v", err)
		return &types.Response{
			Code:    response.ServerErrorCode,
			Message: "Database error",
			Info:    err.Error(),
		}, nil
	}

	return &types.Response{
		Code:    response.SuccessCode,
		Message: "success",
	}, nil
}

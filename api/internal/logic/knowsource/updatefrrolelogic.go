package knowsource

import (
	"context"
	"strings"

	"knowsource/api/internal/svc"
	"knowsource/api/internal/types"
	"knowsource/common/response"
	"knowsource/model"

	"github.com/pkg/errors"
	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateFrRoleLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 更新角色
func NewUpdateFrRoleLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateFrRoleLogic {
	return &UpdateFrRoleLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UpdateFrRoleLogic) UpdateFrRole(req *types.FrRoleUpdateRequest) (resp *types.Response, err error) {
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

	// 检查角色是否存在
	existingRole, err := l.svcCtx.FrRolesModel.FindOneByClientIdRole(l.ctx, clientId, req.Role)
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

	// 更新角色
	role := &model.FrRoles{
		Id:       existingRole.Id,
		ClientId: clientId,
		Role:     req.Role,
		Name:     req.Name,
	}

	err = l.svcCtx.FrRolesModel.Update(l.ctx, role)
	if err != nil {
		l.Logger.Errorf("更新角色失败: %v", err)
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

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

type UpdateFrUserRoleLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 更新用户角色关联
func NewUpdateFrUserRoleLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateFrUserRoleLogic {
	return &UpdateFrUserRoleLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UpdateFrUserRoleLogic) UpdateFrUserRole(req *types.FrUserRoleUpdateRequest) (resp *types.Response, err error) {
	if req.Id == 0 {
		return &types.Response{
			Code:    response.ParameterErrorCode,
			Message: "关联ID不能为空",
		}, nil
	}

	if req.EmpCode == "" {
		return &types.Response{
			Code:    response.ParameterErrorCode,
			Message: "员工编码不能为空",
		}, nil
	}

	if req.Role == "" {
		return &types.Response{
			Code:    response.ParameterErrorCode,
			Message: "角色编码不能为空",
		}, nil
	}

	clientId, _ := l.ctx.Value("clientId").(string)
	if strings.TrimSpace(clientId) == "" {
		return &types.Response{
			Code:    response.UnauthorizedCode,
			Message: "clientId不能为空，请重新登录",
		}, nil
	}

	// 检查关联是否存在
	_, err = l.svcCtx.FrUserRolesModel.FindOne(l.ctx, req.Id)
	if err != nil {
		if errors.Is(err, model.ErrNotFound) {
			return &types.Response{
				Code:    response.NotFoundCode,
				Message: "用户角色关联不存在",
			}, nil
		}
		l.Logger.Errorf("查询用户角色关联失败: %v", err)
		return &types.Response{
			Code:    response.ServerErrorCode,
			Message: "Database error",
			Info:    err.Error(),
		}, nil
	}

	// 检查是否已存在相同的员工和角色组合（排除当前记录）
	existing, err := l.svcCtx.FrUserRolesModel.FindOneByClientIdEmpCodeRole(l.ctx, clientId, req.EmpCode, req.Role)
	if err == nil && existing.Id != req.Id {
		return &types.Response{
			Code:    response.ParameterErrorCode,
			Message: "该员工和角色组合已存在",
		}, nil
	}

	if err != nil && !errors.Is(err, model.ErrNotFound) {
		l.Logger.Errorf("查询用户角色关联失败: %v", err)
		return &types.Response{
			Code:    response.ServerErrorCode,
			Message: "Database error",
			Info:    err.Error(),
		}, nil
	}

	// 更新关联
	userRole := &model.FrUserRoles{
		ClientId: clientId,
		Id:       req.Id,
		EmpCode:  req.EmpCode,
		Role:     req.Role,
	}

	err = l.svcCtx.FrUserRolesModel.Update(l.ctx, userRole)
	if err != nil {
		l.Logger.Errorf("更新用户角色关联失败: %v", err)
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

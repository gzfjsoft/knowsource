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

type CreateFrUserRoleLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 创建用户角色关联
func NewCreateFrUserRoleLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateFrUserRoleLogic {
	return &CreateFrUserRoleLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CreateFrUserRoleLogic) CreateFrUserRole(req *types.FrUserRoleCreateRequest) (resp *types.Response, err error) {
	clientId, _ := l.ctx.Value("clientId").(string)
	if strings.TrimSpace(clientId) == "" {
		return &types.Response{
			Code:    response.UnauthorizedCode,
			Message: "clientId不能为空，请重新登录",
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

	// 检查是否已存在相同的员工和角色组合
	_, err = l.svcCtx.FrUserRolesModel.FindOneByClientIdEmpCodeRole(l.ctx, clientId, req.EmpCode, req.Role)
	if err == nil {
		return &types.Response{
			Code:    response.ParameterErrorCode,
			Message: "该员工和角色组合已存在",
		}, nil
	}

	if !errors.Is(err, model.ErrNotFound) {
		l.Logger.Errorf("查询用户角色关联失败: %v", err)
		return &types.Response{
			Code:    response.ServerErrorCode,
			Message: "Database error",
			Info:    err.Error(),
		}, nil
	}

	// 创建新关联
	userRole := &model.FrUserRoles{
		ClientId: clientId,
		EmpCode:  req.EmpCode,
		Role:     req.Role,
	}

	_, err = l.svcCtx.FrUserRolesModel.Insert(l.ctx, userRole)
	if err != nil {
		l.Logger.Errorf("创建用户角色关联失败: %v", err)
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

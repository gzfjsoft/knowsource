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

type CreateFrRoleLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 创建角色
func NewCreateFrRoleLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateFrRoleLogic {
	return &CreateFrRoleLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CreateFrRoleLogic) CreateFrRole(req *types.FrRoleCreateRequest) (resp *types.Response, err error) {
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

	// 检查是否已存在相同的角色编码
	_, err = l.svcCtx.FrRolesModel.FindOneByClientIdRole(l.ctx, clientId, req.Role)
	if err == nil {
		return &types.Response{
			Code:    response.ParameterErrorCode,
			Message: "该角色编码已存在",
		}, nil
	}

	if !errors.Is(err, model.ErrNotFound) {
		l.Logger.Errorf("查询角色失败: %v", err)
		return &types.Response{
			Code:    response.ServerErrorCode,
			Message: "Database error",
			Info:    err.Error(),
		}, nil
	}

	// 创建新角色
	role := &model.FrRoles{
		ClientId: clientId,
		Role:     req.Role,
		Name:     req.Name,
	}

	_, err = l.svcCtx.FrRolesModel.Insert(l.ctx, role)
	if err != nil {
		l.Logger.Errorf("创建角色失败: %v", err)
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

package knowsource

import (
	"context"

	"knowsource/api/internal/svc"
	"knowsource/api/internal/types"
	"knowsource/common/response"
	"knowsource/model"

	"github.com/pkg/errors"
	"github.com/zeromicro/go-zero/core/logx"
)

type GetFrUserRoleLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 获取用户角色关联
func NewGetFrUserRoleLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetFrUserRoleLogic {
	return &GetFrUserRoleLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetFrUserRoleLogic) GetFrUserRole(req *types.FrUserRoleGetRequest) (resp *types.FrUserRoleGetResponse, err error) {
	if req.Id == 0 {
		return &types.FrUserRoleGetResponse{
			Response: types.Response{
				Code:    response.ParameterErrorCode,
				Message: "关联ID不能为空",
			},
		}, nil
	}

	userRole, err := l.svcCtx.FrUserRolesModel.FindOne(l.ctx, req.Id)
	if err != nil {
		if errors.Is(err, model.ErrNotFound) {
			return &types.FrUserRoleGetResponse{
				Response: types.Response{
					Code:    response.NotFoundCode,
					Message: "用户角色关联不存在",
				},
			}, nil
		}
		l.Logger.Errorf("查询用户角色关联失败: %v", err)
		return &types.FrUserRoleGetResponse{
			Response: types.Response{
				Code:    response.ServerErrorCode,
				Message: "Database error",
				Info:    err.Error(),
			},
		}, nil
	}

	return &types.FrUserRoleGetResponse{
		Response: types.Response{
			Code:    response.SuccessCode,
			Message: "success",
		},
		Data: &types.FrUserRoleInfo{
			Id:      userRole.Id,
			EmpCode: userRole.EmpCode,
			Role:    userRole.Role,
		},
	}, nil
}

package knowsource

import (
	"context"
	"strings"

	"knowsource/api/internal/svc"
	"knowsource/api/internal/types"
	"knowsource/common/response"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListFrUserRoleLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 获取用户角色关联列表
func NewListFrUserRoleLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListFrUserRoleLogic {
	return &ListFrUserRoleLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ListFrUserRoleLogic) ListFrUserRole(req *types.FrUserRoleListRequest) (resp *types.FrUserRoleListResponse, err error) {
	clientId, _ := l.ctx.Value("clientId").(string)
	clientId = strings.TrimSpace(clientId)
	if clientId == "" {
		return &types.FrUserRoleListResponse{
			Response: types.Response{
				Code:    response.UnauthorizedCode,
				Message: "clientId不能为空，请重新登录",
			},
		}, nil
	}

	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 10
	}

	total, err := l.svcCtx.FrUserRolesModel.CountListByClientId(l.ctx, clientId, req.EmpCode, req.Role)
	if err != nil {
		l.Logger.Errorf("获取用户角色关联总数失败: %v", err)
		return &types.FrUserRoleListResponse{
			Response: types.Response{
				Code:    response.ServerErrorCode,
				Message: "获取用户角色关联总数失败",
				Info:    err.Error(),
			},
		}, nil
	}

	offset := (req.Page - 1) * req.PageSize
	rows, err := l.svcCtx.FrUserRolesModel.FindListByClientId(l.ctx, clientId, req.EmpCode, req.Role, int64(req.PageSize), int64(offset))
	if err != nil {
		l.Logger.Errorf("获取用户角色关联列表失败: %v", err)
		return &types.FrUserRoleListResponse{
			Response: types.Response{
				Code:    response.ServerErrorCode,
				Message: "获取用户角色关联列表失败",
				Info:    err.Error(),
			},
		}, nil
	}

	list := make([]types.FrUserRoleInfo, len(rows))
	for i, row := range rows {
		list[i] = types.FrUserRoleInfo{
			Id:      row.Id,
			EmpCode: row.EmpCode,
			Role:    row.Role,
		}
	}

	return &types.FrUserRoleListResponse{
		Response: types.Response{
			Code:    response.SuccessCode,
			Message: "success",
		},
		Data: &types.FrUserRoleListData{
			List:  list,
			Total: total,
		},
	}, nil
}

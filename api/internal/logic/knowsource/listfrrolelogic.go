package knowsource

import (
	"context"
	"fmt"
	"strings"

	"knowsource/api/internal/svc"
	"knowsource/api/internal/types"
	"knowsource/common/response"
	"knowsource/model"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type ListFrRoleLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 获取角色列表
func NewListFrRoleLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListFrRoleLogic {
	return &ListFrRoleLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ListFrRoleLogic) ListFrRole(req *types.FrRoleListRequest) (resp *types.FrRoleListResponse, err error) {
	clientId, _ := l.ctx.Value("clientId").(string)
	clientId = strings.TrimSpace(clientId)
	if clientId == "" {
		return &types.FrRoleListResponse{
			Response: types.Response{
				Code:    response.UnauthorizedCode,
				Message: "clientId不能为空，请重新登录",
			},
		}, nil
	}

	// 设置默认分页参数
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 10
	}

	// 构建查询条件
	var conditions []string
	var args []interface{}
	conditions = append(conditions, "client_id = ?")
	args = append(args, clientId)

	if req.Role != "" {
		conditions = append(conditions, "role = ?")
		args = append(args, req.Role)
	}

	if req.Name != "" {
		conditions = append(conditions, "name LIKE ?")
		args = append(args, "%"+req.Name+"%")
	}

	whereClause := ""
	if len(conditions) > 0 {
		whereClause = "WHERE " + strings.Join(conditions, " AND ")
	}

	// 查询总数
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM fr_roles %s", whereClause)
	var total int64
	err = l.svcCtx.Mysql.QueryRowCtx(l.ctx, &total, countQuery, args...)
	if err != nil {
		l.Logger.Errorf("获取角色总数失败: %v", err)
		return &types.FrRoleListResponse{
			Response: types.Response{
				Code:    response.ServerErrorCode,
				Message: "获取角色总数失败",
				Info:    err.Error(),
			},
		}, nil
	}

	// 查询列表数据
	offset := (req.Page - 1) * req.PageSize
	query := fmt.Sprintf(`
		SELECT 
			id,
			client_id,
			role,
			name
		FROM fr_roles
		%s
		ORDER BY role
		LIMIT ? OFFSET ?
	`, whereClause)
	args = append(args, req.PageSize, offset)

	var rows []model.FrRoles
	err = l.svcCtx.Mysql.QueryRowsCtx(l.ctx, &rows, query, args...)
	if err != nil {
		if err == sqlx.ErrNotFound {
			return &types.FrRoleListResponse{
				Response: types.Response{
					Code:    response.SuccessCode,
					Message: "查询成功",
				},
				Data: &types.FrRoleListData{
					List:  []types.FrRoleInfo{},
					Total: 0,
				},
			}, nil
		}
		l.Logger.Errorf("获取角色列表失败: %v", err)
		return &types.FrRoleListResponse{
			Response: types.Response{
				Code:    response.ServerErrorCode,
				Message: "获取角色列表失败",
				Info:    err.Error(),
			},
		}, nil
	}

	// 转换为响应格式
	list := make([]types.FrRoleInfo, len(rows))
	for i, row := range rows {
		list[i] = types.FrRoleInfo{
			Role: row.Role,
			Name: row.Name,
		}
	}

	return &types.FrRoleListResponse{
		Response: types.Response{
			Code:    response.SuccessCode,
			Message: "success",
		},
		Data: &types.FrRoleListData{
			List:  list,
			Total: total,
		},
	}, nil
}

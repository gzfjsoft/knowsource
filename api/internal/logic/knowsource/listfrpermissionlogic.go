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

type ListFrPermissionLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 获取权限列表
func NewListFrPermissionLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListFrPermissionLogic {
	return &ListFrPermissionLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ListFrPermissionLogic) ListFrPermission(req *types.FrPermissionListRequest) (resp *types.FrPermissionListResponse, err error) {
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

	if req.Permission != "" {
		conditions = append(conditions, "permission = ?")
		args = append(args, req.Permission)
	}

	if req.Description != "" {
		conditions = append(conditions, "description LIKE ?")
		args = append(args, "%"+req.Description+"%")
	}

	whereClause := ""
	if len(conditions) > 0 {
		whereClause = "WHERE " + strings.Join(conditions, " AND ")
	}

	// 查询总数
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM fr_permissions %s", whereClause)
	var total int64
	err = l.svcCtx.Mysql.QueryRowCtx(l.ctx, &total, countQuery, args...)
	if err != nil {
		l.Logger.Errorf("获取权限总数失败: %v", err)
		return &types.FrPermissionListResponse{
			Response: types.Response{
				Code:    response.ServerErrorCode,
				Message: "获取权限总数失败",
				Info:    err.Error(),
			},
		}, nil
	}

	// 查询列表数据
	offset := (req.Page - 1) * req.PageSize
	query := fmt.Sprintf(`
		SELECT 
			id,
			permission,
			description
		FROM fr_permissions
		%s
		ORDER BY id ASC
		LIMIT ? OFFSET ?
	`, whereClause)
	args = append(args, req.PageSize, offset)

	var rows []model.FrPermissions
	err = l.svcCtx.Mysql.QueryRowsCtx(l.ctx, &rows, query, args...)
	if err != nil {
		if err == sqlx.ErrNotFound {
			return &types.FrPermissionListResponse{
				Response: types.Response{
					Code:    response.SuccessCode,
					Message: "查询成功",
				},
				Data: &types.FrPermissionListData{
					List:  []types.FrPermissionInfo{},
					Total: 0,
				},
			}, nil
		}
		l.Logger.Errorf("获取权限列表失败: %v", err)
		return &types.FrPermissionListResponse{
			Response: types.Response{
				Code:    response.ServerErrorCode,
				Message: "获取权限列表失败",
				Info:    err.Error(),
			},
		}, nil
	}

	// 转换为响应格式
	list := make([]types.FrPermissionInfo, len(rows))
	for i, row := range rows {
		list[i] = types.FrPermissionInfo{
			Permission:  row.Permission,
			Description: row.Description,
		}
	}

	return &types.FrPermissionListResponse{
		Response: types.Response{
			Code:    response.SuccessCode,
			Message: "success",
		},
		Data: &types.FrPermissionListData{
			List:  list,
			Total: total,
		},
	}, nil
}

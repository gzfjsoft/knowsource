package knowsource

import (
	"context"
	"fmt"
	"strings"

	"knowsource/api/internal/svc"
	"knowsource/api/internal/types"
	"knowsource/common/response"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type ListDeptLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 获取部门列表
func NewListDeptLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListDeptLogic {
	return &ListDeptLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ListDeptLogic) ListDept(req *types.KnowsourceDeptListRequest) (resp *types.KnowsourceDeptListResponse, err error) {
	clientId, _ := l.ctx.Value("clientId").(string)
	clientId = strings.TrimSpace(clientId)
	if clientId == "" {
		return &types.KnowsourceDeptListResponse{
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

	if req.DeptCode != "" {
		conditions = append(conditions, "dept_code = ?")
		args = append(args, req.DeptCode)
	}

	if req.DeptName != "" {
		conditions = append(conditions, "dept_name LIKE ?")
		args = append(args, "%"+req.DeptName+"%")
	}

	if req.ParentCode != "" {
		conditions = append(conditions, "parent_code = ?")
		args = append(args, req.ParentCode)
	}

	if req.Grade > 0 {
		conditions = append(conditions, "GRADE = ?")
		args = append(args, req.Grade)
	}

	whereClause := ""
	if len(conditions) > 0 {
		whereClause = "WHERE " + strings.Join(conditions, " AND ")
	}

	// 查询总数
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM fr_dept %s", whereClause)
	var total int64
	err = l.svcCtx.Mysql.QueryRowCtx(l.ctx, &total, countQuery, args...)
	if err != nil {
		l.Logger.Errorf("获取部门总数失败: %v", err)
		return &types.KnowsourceDeptListResponse{
			Response: types.Response{
				Code:    response.ServerErrorCode,
				Message: "获取部门总数失败",
				Info:    err.Error(),
			},
		}, nil
	}

	// 查询列表数据
	offset := (req.Page - 1) * req.PageSize
	query := fmt.Sprintf(`
		SELECT 
			dept_code,
			dept_name,
			parent_code,
			 
			grade,
			end_mark,
			kind,
			b0110
		FROM fr_dept
		%s
		ORDER BY grade, dept_code
		LIMIT ? OFFSET ?
	`, whereClause)
	args = append(args, req.PageSize, offset)

	type ResultRow struct {
		DeptCode   string `db:"dept_code"`
		DeptName   string `db:"dept_name"`
		ParentCode string `db:"parent_code"`

		Grade   int64  `db:"grade"`
		EndMark string `db:"end_mark"`
		Kind    string `db:"kind"`
		B0110   string `db:"b0110"`
	}

	var rows []ResultRow
	err = l.svcCtx.Mysql.QueryRowsCtx(l.ctx, &rows, query, args...)
	if err != nil {
		if err == sqlx.ErrNotFound {
			return &types.KnowsourceDeptListResponse{
				Response: types.Response{
					Code:    response.SuccessCode,
					Message: "查询成功",
				},
				Data: &types.KnowsourceDeptListData{
					List:  []types.KnowsourceDeptListItem{},
					Total: 0,
				},
			}, nil
		}
		l.Logger.Errorf("获取部门列表失败: %v", err)
		return &types.KnowsourceDeptListResponse{
			Response: types.Response{
				Code:    response.ServerErrorCode,
				Message: "获取部门列表失败",
				Info:    err.Error(),
			},
		}, nil
	}

	// 转换为响应格式
	list := make([]types.KnowsourceDeptListItem, len(rows))
	for i, row := range rows {
		list[i] = types.KnowsourceDeptListItem{
			DeptCode:   row.DeptCode,
			DeptName:   row.DeptName,
			ParentCode: row.ParentCode,
			Grade:      row.Grade,
			EndMark:    row.EndMark,
			Kind:       row.Kind,
			B0110:      row.B0110,
		}
	}

	return &types.KnowsourceDeptListResponse{
		Response: types.Response{
			Code:    response.SuccessCode,
			Message: "查询成功",
		},
		Data: &types.KnowsourceDeptListData{
			List:  list,
			Total: total,
		},
	}, nil
}

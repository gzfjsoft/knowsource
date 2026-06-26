package knowsource

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"
	"strings"

	"knowsource/api/internal/svc"
	"knowsource/api/internal/types"
	"knowsource/common/response"
	"knowsource/model"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type ListEmpLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 获取员工列表
func NewListEmpLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListEmpLogic {
	return &ListEmpLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ListEmpLogic) ListEmp(req *types.KnowsourceEmpListRequest) (resp *types.KnowsourceEmpListResponse, err error) {
	clientId, _ := l.ctx.Value("clientId").(string)
	clientId = strings.TrimSpace(clientId)
	if clientId == "" {
		return &types.KnowsourceEmpListResponse{
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

	logx.Infof("req: %v", req)

	// 构建查询条件
	var conditions []string
	var args []interface{}

	conditions = append(conditions, "e.client_id = ?")
	args = append(args, clientId)

	if req.EmpCode != "" {
		conditions = append(conditions, "e.femp_code = ?")
		args = append(args, req.EmpCode)
	}

	if req.EmpName != "" {
		conditions = append(conditions, "e.femp_name LIKE ?")
		args = append(args, "%"+req.EmpName+"%")
	}

	if req.DeptCode != "" {
		conditions = append(conditions, "e.dept_code LIKE ?")
		args = append(args, req.DeptCode+"%")
	}

	// 状态筛选：空串忽略，"0"查询0，"1"查询1
	if req.Status != "" {
		if req.Status == "0" || req.Status == "1" {
			statusInt, err := strconv.ParseInt(req.Status, 10, 64)
			if err == nil {
				conditions = append(conditions, "e.status = ?")
				args = append(args, statusInt)
			}
		}
	}

	whereClause := ""
	if len(conditions) > 0 {
		whereClause = "WHERE " + strings.Join(conditions, " AND ")
	}

	logx.Infof("whereClause: %s", whereClause)
	logx.Infof("args: %v", args)

	// 查询总数
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM fr_emp e %s", whereClause)
	var total int64
	err = l.svcCtx.Mysql.QueryRowCtx(l.ctx, &total, countQuery, args...)
	if err != nil {
		l.Logger.Errorf("获取员工总数失败: %v", err)
		return &types.KnowsourceEmpListResponse{
			Response: types.Response{
				Code:    response.ServerErrorCode,
				Message: "获取员工总数失败",
				Info:    err.Error(),
			},
		}, nil
	}

	// 查询列表数据
	offset := (req.Page - 1) * req.PageSize
	query := fmt.Sprintf(`
		SELECT 
			e.femp_code as FempCode,
			e.femp_name as FempName,
			e.dept_code as DeptCode,
			d.dept_name as DeptName,

			e.fposition as Fposition,
			e.status,

			e.fbranch as FBranch
		FROM fr_emp e
		LEFT JOIN fr_dept d ON e.dept_code = d.dept_code AND d.client_id = e.client_id
		%s
		ORDER BY e.femp_code
		LIMIT ? OFFSET ?
	`, whereClause)
	args = append(args, req.PageSize, offset)

	type ResultRow struct {
		FempCode string         `db:"FempCode"`
		FempName string         `db:"FempName"`
		DeptCode string         `db:"DeptCode"`
		DeptName sql.NullString `db:"DeptName"`

		Fposition string `db:"Fposition"`
		Status    int64  `db:"status"`

		FBranch string `db:"FBranch"`
	}

	var rows []ResultRow
	err = l.svcCtx.Mysql.QueryRowsCtx(l.ctx, &rows, query, args...)
	if err != nil {
		if err == sqlx.ErrNotFound {
			return &types.KnowsourceEmpListResponse{
				Response: types.Response{
					Code:    response.SuccessCode,
					Message: "查询成功",
				},
				Data: &types.KnowsourceEmpListData{
					List:  []types.KnowsourceEmpListItem{},
					Total: 0,
				},
			}, nil
		}
		l.Logger.Errorf("获取员工列表失败: %v", err)
		return &types.KnowsourceEmpListResponse{
			Response: types.Response{
				Code:    response.ServerErrorCode,
				Message: "获取员工列表失败",
				Info:    err.Error(),
			},
		}, nil
	}

	// 转换为响应格式，并查询每个员工的角色
	list := make([]types.KnowsourceEmpListItem, len(rows))
	for i, row := range rows {
		// 查询员工的所有角色
		frRols, err := l.svcCtx.FrUserRolesModel.FindAllByClientIdEmpCode(l.ctx, clientId, row.FempCode)
		if err != nil {
			l.Logger.Errorf("查询员工 %s 角色失败: %v", row.FempCode, err)
			// 如果查询角色失败，使用默认角色
			frRols = []*model.FrUserRoles{}
		}

		// 构建角色数组
		roles := make([]string, 0, len(frRols))
		for _, role := range frRols {
			roles = append(roles, role.Role)
		}

		// 如果没有角色，默认添加 "user" 角色
		if len(roles) == 0 {
			roles = []string{"user"}
		}

		// 处理部门名称，如果为空则使用空字符串
		deptName := ""
		if row.DeptName.Valid {
			deptName = row.DeptName.String
		}

		list[i] = types.KnowsourceEmpListItem{
			EmpCode:  row.FempCode,
			EmpName:  row.FempName,
			DeptCode: row.DeptCode,
			DeptName: deptName,
			Position: row.Fposition,
			Status:   row.Status,
			Branch:   row.FBranch,
			Roles:    roles,
		}
	}

	return &types.KnowsourceEmpListResponse{
		Response: types.Response{
			Code:    response.SuccessCode,
			Message: "查询成功",
		},
		Data: &types.KnowsourceEmpListData{
			List:  list,
			Total: total,
		},
	}, nil
}

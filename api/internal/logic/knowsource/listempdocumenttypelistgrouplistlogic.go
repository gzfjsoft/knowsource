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

type ListEmpDocumentTypeListGroupListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 获取员工知识库权限列表group
func NewListEmpDocumentTypeListGroupListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListEmpDocumentTypeListGroupListLogic {
	return &ListEmpDocumentTypeListGroupListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ListEmpDocumentTypeListGroupListLogic) ListEmpDocumentTypeListGroupList(req *types.KnowsourceEmpDocumentTypeListRequest) (resp *types.KnowsourceEmpDocumentTypeListGroupResponse, err error) {
	// 设置默认分页参数
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 10
	}

	// 获取租户ID
	clientId, _ := l.ctx.Value("clientId").(string)

	// 构建查询条件
	var conditions []string
	var args []interface{}

	// 添加租户隔离
	conditions = append(conditions, "edt.client_id = ?")
	args = append(args, clientId)

	if req.EmpCode != "" {
		conditions = append(conditions, "edt.emp_code = ?")
		args = append(args, req.EmpCode)
	}

	if req.EmpName != "" {
		conditions = append(conditions, "emp.femp_name LIKE ?")
		args = append(args, "%"+req.EmpName+"%")
	}

	whereClause := ""
	if len(conditions) > 0 {
		whereClause = "WHERE " + strings.Join(conditions, " AND ")
	}

	// 查询总数（按员工分组后的数量）
	// 如果条件中使用了 emp.FempName，需要 JOIN fr_emp 表
	countJoinClause := ""
	if req.EmpName != "" {
		countJoinClause = "LEFT JOIN fr_emp emp ON edt.emp_code = emp.femp_code AND edt.client_id = emp.client_id"
	}
	countQuery := fmt.Sprintf(`
		SELECT COUNT(DISTINCT edt.emp_code) 
		FROM emp_document_type edt
		%s
		%s
	`, countJoinClause, whereClause)
	var total int64
	err = l.svcCtx.Mysql.QueryRowCtx(l.ctx, &total, countQuery, args...)
	if err != nil {
		return &types.KnowsourceEmpDocumentTypeListGroupResponse{
			Response: types.Response{
				Code:    response.ServerErrorCode,
				Message: "查询总数失败",
				Info:    err.Error(),
			},
		}, nil
	}

	// 查询分组数据：员工部门编码前三位 JOIN fr_dept 得到公司名称（CONTENT）
	offset := (req.Page - 1) * req.PageSize
	query := fmt.Sprintf(`
		SELECT 
			edt.emp_code,
			COALESCE(emp.femp_name, '') as emp_name,
			COALESCE(emp.dept_code, '') as dept_code,
			COALESCE(dept.dept_name, '') as dept_name,
			COALESCE(company_dept.dept_name, '') as company_name,
			GROUP_CONCAT(DISTINCT edt.document_type_code ORDER BY edt.document_type_code SEPARATOR ',') as document_type_codes,
			GROUP_CONCAT(DISTINCT dt.name ORDER BY edt.document_type_code SEPARATOR ',') as document_type_names
		FROM emp_document_type edt
		LEFT JOIN fr_emp emp ON edt.emp_code = emp.femp_code AND edt.client_id = emp.client_id
		LEFT JOIN fr_dept dept ON emp.dept_code = dept.dept_code AND emp.client_id = dept.client_id
		LEFT JOIN fr_dept company_dept ON company_dept.dept_code = LEFT(emp.dept_code, 3) AND emp.client_id = company_dept.client_id
		LEFT JOIN document_type dt ON edt.document_type_code = dt.code AND edt.client_id = dt.client_id
		%s
		GROUP BY edt.emp_code, emp.femp_name, emp.dept_code, dept.dept_name, company_dept.dept_name
		ORDER BY edt.emp_code
		LIMIT ? OFFSET ?
	`, whereClause)
	args = append(args, req.PageSize, offset)

	type ResultRow struct {
		EmpCode           string `db:"emp_code"`
		EmpName           string `db:"emp_name"`
		DeptCode          string `db:"dept_code"`
		DeptName          string `db:"dept_name"`
		CompanyName       string `db:"company_name"`
		DocumentTypeCodes string `db:"document_type_codes"`
		DocumentTypeNames string `db:"document_type_names"`
	}

	var rows []ResultRow
	err = l.svcCtx.Mysql.QueryRowsCtx(l.ctx, &rows, query, args...)
	if err != nil {
		if err == sqlx.ErrNotFound {
			return &types.KnowsourceEmpDocumentTypeListGroupResponse{
				Response: types.Response{
					Code:    response.SuccessCode,
					Message: "查询成功",
				},
				Data: &types.KnowsourceEmpDocumentTypeListGroupData{
					List:  []types.KnowsourceEmpDocumentTypeInfoGroup{},
					Total: 0,
				},
			}, nil
		}
		return &types.KnowsourceEmpDocumentTypeListGroupResponse{
			Response: types.Response{
				Code:    response.ServerErrorCode,
				Message: "查询失败",
				Info:    err.Error(),
			},
		}, nil
	}

	// 转换为响应格式
	list := make([]types.KnowsourceEmpDocumentTypeInfoGroup, len(rows))
	for i, row := range rows {
		// 将逗号分隔的字符串转换为数组
		var documentTypeCodes []string
		var documentTypeNames []string

		if row.DocumentTypeCodes != "" {
			documentTypeCodes = strings.Split(row.DocumentTypeCodes, ",")
		} else {
			documentTypeCodes = []string{}
		}

		if row.DocumentTypeNames != "" {
			documentTypeNames = strings.Split(row.DocumentTypeNames, ",")
		} else {
			documentTypeNames = []string{}
		}

		list[i] = types.KnowsourceEmpDocumentTypeInfoGroup{
			EmpCode:           row.EmpCode,
			EmpName:           row.EmpName,
			DeptCode:          row.DeptCode,
			DeptName:          row.DeptName,
			CompanyName:       row.CompanyName,
			DocumentTypeCodes: documentTypeCodes,
			DocumentTypeNames: documentTypeNames,
		}
	}

	return &types.KnowsourceEmpDocumentTypeListGroupResponse{
		Response: types.Response{
			Code:    response.SuccessCode,
			Message: "查询成功",
		},
		Data: &types.KnowsourceEmpDocumentTypeListGroupData{
			List:  list,
			Total: total,
		},
	}, nil
}

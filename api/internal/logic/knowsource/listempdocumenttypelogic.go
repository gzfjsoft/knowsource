package knowsource

import (
	"context"
	"fmt"
	"strings"
	"time"

	"knowsource/api/internal/svc"
	"knowsource/api/internal/types"
	"knowsource/common/response"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type ListEmpDocumentTypeLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 获取员工知识库权限列表
func NewListEmpDocumentTypeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListEmpDocumentTypeLogic {
	return &ListEmpDocumentTypeLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ListEmpDocumentTypeLogic) ListEmpDocumentType(req *types.KnowsourceEmpDocumentTypeListRequest) (resp *types.KnowsourceEmpDocumentTypeListResponse, err error) {
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
		conditions = append(conditions, "emp.femp_name like ?")
		args = append(args, "%"+req.EmpName+"%")
	}

	whereClause := ""
	if len(conditions) > 0 {
		whereClause = "WHERE " + strings.Join(conditions, " AND ")
	}

	// 查询总数
	countJoinClause := ""
	if req.EmpName != "" {
		countJoinClause = "LEFT JOIN fr_emp emp ON edt.emp_code = emp.femp_code AND edt.client_id = emp.client_id"
	}
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM emp_document_type edt %s %s", countJoinClause, whereClause)
	var total int64
	err = l.svcCtx.Mysql.QueryRowCtx(l.ctx, &total, countQuery, args...)
	if err != nil {
		return &types.KnowsourceEmpDocumentTypeListResponse{
			Response: types.Response{
				Code:    response.ServerErrorCode,
				Message: err.Error(),
			},
		}, nil
	}

	// 查询列表数据
	offset := (req.Page - 1) * req.PageSize
	query := fmt.Sprintf(`
		SELECT 
			edt.id,
			edt.emp_code,
			COALESCE(emp.femp_name, '') as emp_name,
			edt.document_type_code,
			COALESCE(dt.name, '') as document_type_name,
			edt.created_at,
			edt.updated_at
		FROM emp_document_type edt
		LEFT JOIN fr_emp emp ON edt.emp_code = emp.femp_code AND edt.client_id = emp.client_id
		LEFT JOIN document_type dt ON edt.document_type_code = dt.code AND edt.client_id = dt.client_id
		%s
		ORDER BY edt.created_at DESC
		LIMIT ? OFFSET ?
	`, whereClause)
	args = append(args, req.PageSize, offset)

	type ResultRow struct {
		Id               int64     `db:"id"`
		EmpCode          string    `db:"emp_code"`
		EmpName          string    `db:"emp_name"`
		DocumentTypeCode string    `db:"document_type_code"`
		DocumentTypeName string    `db:"document_type_name"`
		CreatedAt        time.Time `db:"created_at"`
		UpdatedAt        time.Time `db:"updated_at"`
	}

	var rows []ResultRow
	err = l.svcCtx.Mysql.QueryRowsCtx(l.ctx, &rows, query, args...)
	if err != nil {
		if err == sqlx.ErrNotFound {
			return &types.KnowsourceEmpDocumentTypeListResponse{
				Response: types.Response{
					Code:    response.SuccessCode,
					Message: "查询成功",
				},
				Data: &types.KnowsourceEmpDocumentTypeListData{
					List:  []types.KnowsourceEmpDocumentTypeInfo{},
					Total: 0,
				},
			}, nil
		}
		return &types.KnowsourceEmpDocumentTypeListResponse{
			Response: types.Response{
				Code:    response.ServerErrorCode,
				Message: err.Error(),
			},
		}, nil
	}

	// 转换为响应格式
	list := make([]types.KnowsourceEmpDocumentTypeInfo, len(rows))
	for i, row := range rows {
		list[i] = types.KnowsourceEmpDocumentTypeInfo{
			Id:               row.Id,
			EmpCode:          row.EmpCode,
			EmpName:          row.EmpName,
			DocumentTypeCode: row.DocumentTypeCode,
			DocumentTypeName: row.DocumentTypeName,
			CreatedAt:        row.CreatedAt.Format("2006-01-02 15:04:05"),
			UpdatedAt:        row.UpdatedAt.Format("2006-01-02 15:04:05"),
		}
	}

	return &types.KnowsourceEmpDocumentTypeListResponse{
		Response: types.Response{
			Code:    response.SuccessCode,
			Message: "查询成功",
		},
		Data: &types.KnowsourceEmpDocumentTypeListData{
			List:  list,
			Total: total,
		},
	}, nil
}

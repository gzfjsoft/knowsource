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

type ListDeptDocumentTypeLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 获取部门文档类型绑定列表
func NewListDeptDocumentTypeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListDeptDocumentTypeLogic {
	return &ListDeptDocumentTypeLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ListDeptDocumentTypeLogic) ListDeptDocumentType(req *types.KnowsourceDeptDocumentTypeListRequest) (resp *types.KnowsourceDeptDocumentTypeListResponse, err error) {
	clientId, _ := l.ctx.Value("clientId").(string)
	clientId = strings.TrimSpace(clientId)
	if clientId == "" {
		return &types.KnowsourceDeptDocumentTypeListResponse{
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

	conditions = append(conditions, "ddt.client_id = ?")
	args = append(args, clientId)

	if req.DeptCode != "" {
		conditions = append(conditions, "ddt.dept_code = ?")
		args = append(args, req.DeptCode)
	}

	if req.DocumentTypeCode != "" {
		conditions = append(conditions, "ddt.document_type_code = ?")
		args = append(args, req.DocumentTypeCode)
	}

	whereClause := ""
	if len(conditions) > 0 {
		whereClause = "WHERE " + strings.Join(conditions, " AND ")
	}

	// 查询总数
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM dept_document_type ddt %s", whereClause)
	var total int64
	err = l.svcCtx.Mysql.QueryRowCtx(l.ctx, &total, countQuery, args...)
	if err != nil {
		return &types.KnowsourceDeptDocumentTypeListResponse{
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
			ddt.id,
			ddt.dept_code,
			COALESCE(dept.dept_name, '') as dept_name,
			ddt.document_type_code,
			COALESCE(dt.name, '') as document_type_name,
			ddt.created_at,
			ddt.updated_at
		FROM dept_document_type ddt
		LEFT JOIN fr_dept dept ON ddt.dept_code = dept.dept_code AND ddt.client_id = dept.client_id
		LEFT JOIN document_type dt ON ddt.document_type_code = dt.code AND ddt.client_id = dt.client_id
		%s
		ORDER BY ddt.created_at DESC
		LIMIT ? OFFSET ?
	`, whereClause)
	args = append(args, req.PageSize, offset)

	type ResultRow struct {
		Id               int64     `db:"id"`
		DeptCode         string    `db:"dept_code"`
		DeptName         string    `db:"dept_name"`
		DocumentTypeCode string    `db:"document_type_code"`
		DocumentTypeName string    `db:"document_type_name"`
		CreatedAt        time.Time `db:"created_at"`
		UpdatedAt        time.Time `db:"updated_at"`
	}

	var rows []ResultRow
	err = l.svcCtx.Mysql.QueryRowsCtx(l.ctx, &rows, query, args...)
	if err != nil {
		if err == sqlx.ErrNotFound {
			return &types.KnowsourceDeptDocumentTypeListResponse{
				Response: types.Response{
					Code:    response.SuccessCode,
					Message: "查询成功",
				},
				Data: &types.KnowsourceDeptDocumentTypeListData{
					List:  []types.KnowsourceDeptDocumentTypeInfo{},
					Total: 0,
				},
			}, nil
		}
		return &types.KnowsourceDeptDocumentTypeListResponse{
			Response: types.Response{
				Code:    response.ServerErrorCode,
				Message: err.Error(),
			},
		}, nil
	}

	// 转换为响应格式
	list := make([]types.KnowsourceDeptDocumentTypeInfo, len(rows))
	for i, row := range rows {
		list[i] = types.KnowsourceDeptDocumentTypeInfo{
			Id:               row.Id,
			DeptCode:         row.DeptCode,
			DeptName:         row.DeptName,
			DocumentTypeCode: row.DocumentTypeCode,
			DocumentTypeName: row.DocumentTypeName,
			CreatedAt:        row.CreatedAt.Format("2006-01-02 15:04:05"),
			UpdatedAt:        row.UpdatedAt.Format("2006-01-02 15:04:05"),
		}
	}

	return &types.KnowsourceDeptDocumentTypeListResponse{
		Response: types.Response{
			Code:    response.SuccessCode,
			Message: "查询成功",
		},
		Data: &types.KnowsourceDeptDocumentTypeListData{
			List:  list,
			Total: total,
		},
	}, nil
}

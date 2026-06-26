package knowsource

import (
	"context"
	"fmt"
	"strings"

	"knowsource/api/internal/svc"
	"knowsource/api/internal/types"
	"knowsource/api/internal/utils"
	"knowsource/common/response"
	"knowsource/model"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type ListMyDocumentTypeLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 获取我的文档类型
func NewListMyDocumentTypeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListMyDocumentTypeLogic {
	return &ListMyDocumentTypeLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ListMyDocumentTypeLogic) ListMyDocumentType() (resp *types.KnowsourceMyDocumentTypeListResponse, err error) {
	// 从 context 获取员工编码
	empCodeValue := l.ctx.Value("empCode")
	if empCodeValue == nil {
		return &types.KnowsourceMyDocumentTypeListResponse{
			Response: types.Response{
				Code:    response.ParameterErrorCode,
				Message: "未获取到员工编码",
			},
		}, nil
	}

	empCode, ok := empCodeValue.(string)
	if !ok || empCode == "" {
		return &types.KnowsourceMyDocumentTypeListResponse{
			Response: types.Response{
				Code:    response.ParameterErrorCode,
				Message: "员工编码无效",
			},
		}, nil
	}

	clientId, _ := l.ctx.Value("clientId").(string)
	if strings.TrimSpace(clientId) == "" {
		return &types.KnowsourceMyDocumentTypeListResponse{
			Response: types.Response{
				Code:    response.UnauthorizedCode,
				Message: "clientId不能为空，请重新登录",
			},
		}, nil
	}

	// 查询员工信息
	emp, err := l.svcCtx.FrEmpModel.FindOneByClientIdFempCode(l.ctx, clientId, empCode)
	if err != nil {
		if err == model.ErrNotFound {
			return &types.KnowsourceMyDocumentTypeListResponse{
				Response: types.Response{
					Code:    response.UserNotExistCode,
					Message: "员工不存在",
				},
			}, nil
		}
		return &types.KnowsourceMyDocumentTypeListResponse{
			Response: types.Response{
				Code:    response.ServerErrorCode,
				Message: err.Error(),
			},
		}, nil
	}

	role, _ := utils.GetRoleFromContext(l.ctx)

	// 如果是管理员，直接返回所有文档类型
	if utils.IsAdminRole(role) {
		documentTypes, err := l.svcCtx.DocumentTypeModel.FindAllByClientId(l.ctx, clientId)
		if err != nil {
			l.Logger.Errorf("查询所有文档类型失败: %v", err)
			return &types.KnowsourceMyDocumentTypeListResponse{
				Response: types.Response{
					Code:    response.ServerErrorCode,
					Message: "查询文档类型失败: " + err.Error(),
				},
			}, nil
		}

		// 转换为响应格式
		list := make([]types.KnowsourceMyDocumentTypeInfo, 0, len(documentTypes))
		for _, docType := range documentTypes {
			list = append(list, types.KnowsourceMyDocumentTypeInfo{
				DeptName:         "(全部)",
				DocumentTypeCode: docType.Code,
				DocumentTypeName: docType.Name,
				Description:      docType.Description,
				IsDisabled:       docType.IsDisabled,
			})
		}

		return &types.KnowsourceMyDocumentTypeListResponse{
			Response: types.Response{
				Code:    response.SuccessCode,
				Message: "查询成功",
				Info:    "管理员：返回所有文档类型",
			},
			Data: &types.KnowsourceMyDocumentTypeListData{
				List:  list,
				Total: int64(len(list)),
			},
		}, nil
	}

	// 获取员工所属部门及所有上级部门的合集
	deptCodes := l.getAllParentDepts(l.ctx, emp.DeptCode)
	if len(deptCodes) == 0 && emp.DeptCode != "" {
		deptCodes = []string{emp.DeptCode}
	}

	// 获取部门名称列表用于日志和返回信息
	deptInfoList := l.getDeptInfoList(l.ctx, deptCodes)
	deptInfoStr := l.formatDeptInfo(deptInfoList)

	// 打印日志
	l.Logger.Infof("员工 %s (编码: %s) 所属全部部门: %s", emp.FempName, empCode, deptInfoStr)

	// 用于去重的 map，key 为 DocumentTypeCode
	resultMap := make(map[string]*types.KnowsourceMyDocumentTypeInfo)

	// 1. 查询部门文档类型绑定
	if len(deptCodes) > 0 {
		placeholders := make([]string, len(deptCodes))
		args := make([]interface{}, len(deptCodes))
		for i, deptCode := range deptCodes {
			placeholders[i] = "?"
			args[i] = deptCode
		}

		query := fmt.Sprintf(`
			SELECT 
				COALESCE(dept.dept_name, '') as dept_name,
				ddt.document_type_code,
				COALESCE(dt.name, '') as document_type_name,
				COALESCE(dt.description, '') as description,
				COALESCE(dt.is_disabled, 0) as is_disabled
			FROM dept_document_type ddt
			LEFT JOIN fr_dept dept ON ddt.client_id = dept.client_id AND ddt.dept_code = dept.dept_code
			LEFT JOIN document_type dt ON ddt.client_id = dt.client_id AND ddt.document_type_code = dt.code
			WHERE ddt.client_id = ? AND ddt.dept_code IN (%s)
		`, strings.Join(placeholders, ","))

		type DeptDocRow struct {
			DeptName         string `db:"dept_name"`
			DocumentTypeCode string `db:"document_type_code"`
			DocumentTypeName string `db:"document_type_name"`
			Description      string `db:"description"`
			IsDisabled       int64  `db:"is_disabled"`
		}

		var deptRows []DeptDocRow
		args2 := make([]interface{}, 0, len(args)+1)
		args2 = append(args2, clientId)
		args2 = append(args2, args...)
		err = l.svcCtx.Mysql.QueryRowsCtx(l.ctx, &deptRows, query, args2...)
		if err != nil && err != sqlx.ErrNotFound {
			l.Logger.Errorf("查询部门文档类型失败: %v", err)
		} else {
			for _, row := range deptRows {
				if row.DocumentTypeCode != "" {
					resultMap[row.DocumentTypeCode] = &types.KnowsourceMyDocumentTypeInfo{
						DeptName:         row.DeptName,
						DocumentTypeCode: row.DocumentTypeCode,
						DocumentTypeName: row.DocumentTypeName,
						Description:      row.Description,
						IsDisabled:       row.IsDisabled,
					}
				}
			}
		}
	}

	// 2. 查询员工知识库权限
	empQuery := `
		SELECT 
			edt.document_type_code,
			COALESCE(dt.name, '') as document_type_name,
			COALESCE(dt.description, '') as description,
			COALESCE(dt.is_disabled, 0) as is_disabled
		FROM emp_document_type edt
		LEFT JOIN document_type dt ON edt.client_id = dt.client_id AND edt.document_type_code = dt.code
		WHERE edt.client_id = ? AND edt.emp_code = ?
	`

	type EmpDocRow struct {
		DocumentTypeCode string `db:"document_type_code"`
		DocumentTypeName string `db:"document_type_name"`
		Description      string `db:"description"`
		IsDisabled       int64  `db:"is_disabled"`
	}

	var empRows []EmpDocRow
	err = l.svcCtx.Mysql.QueryRowsCtx(l.ctx, &empRows, empQuery, clientId, empCode)
	if err != nil && err != sqlx.ErrNotFound {
		l.Logger.Errorf("查询员工文档类型失败: %v", err)
	} else {
		for _, row := range empRows {
			if row.DocumentTypeCode != "" {
				// 如果已存在，更新为员工自己的（deptName 为"自己"）
				resultMap[row.DocumentTypeCode] = &types.KnowsourceMyDocumentTypeInfo{
					DeptName:         "(自己)",
					DocumentTypeCode: row.DocumentTypeCode,
					DocumentTypeName: row.DocumentTypeName,
					Description:      row.Description,
					IsDisabled:       row.IsDisabled,
				}
			}
		}
	}

	// 转换为列表
	list := make([]types.KnowsourceMyDocumentTypeInfo, 0, len(resultMap))
	for _, info := range resultMap {
		list = append(list, *info)
	}

	return &types.KnowsourceMyDocumentTypeListResponse{
		Response: types.Response{
			Code:    response.SuccessCode,
			Message: "查询成功",
			Info:    deptInfoStr,
		},
		Data: &types.KnowsourceMyDocumentTypeListData{
			List:  list,
			Total: int64(len(list)),
		},
	}, nil
}

// getAllParentDepts 获取部门及其所有上级部门的编码列表
func (l *ListMyDocumentTypeLogic) getAllParentDepts(ctx context.Context, deptCode string) []string {
	if deptCode == "" {
		return []string{}
	}

	clientId, _ := ctx.Value("clientId").(string)
	if strings.TrimSpace(clientId) == "" {
		return []string{deptCode}
	}

	deptSet := make(map[string]bool)
	deptList := []string{}

	// 递归获取所有上级部门
	var getParents func(code string)
	getParents = func(code string) {
		if code == "" {
			return
		}
		// 避免循环引用
		if deptSet[code] {
			return
		}
		deptSet[code] = true
		deptList = append(deptList, code)

		// 查询当前部门
		dept, err := l.svcCtx.FrDeptModel.FindOneByClientIdDeptCode(ctx, clientId, code)
		if err != nil {
			return
		}

		// 如果有父部门，继续向上查找
		if dept.ParentCode != "" && dept.ParentCode != code {
			getParents(dept.ParentCode)
		}
	}

	getParents(deptCode)
	return deptList
}

// getDeptInfoList 获取部门信息列表（编码和名称）
func (l *ListMyDocumentTypeLogic) getDeptInfoList(ctx context.Context, deptCodes []string) []map[string]string {
	if len(deptCodes) == 0 {
		return []map[string]string{}
	}

	clientId, _ := ctx.Value("clientId").(string)

	deptInfoList := make([]map[string]string, 0, len(deptCodes))
	for _, deptCode := range deptCodes {
		var (
			dept *model.FrDept
			err  error
		)
		if strings.TrimSpace(clientId) != "" {
			dept, err = l.svcCtx.FrDeptModel.FindOneByClientIdDeptCode(ctx, clientId, deptCode)
		} else {
			// 兜底：若无 clientId，则不再查库，避免误返回其他 client 的数据
			err = model.ErrNotFound
		}
		if err != nil {
			// 如果查询失败，仍然记录部门编码
			if deptCode != "ROOT" {
				deptInfoList = append(deptInfoList, map[string]string{
					"code": deptCode,
					"name": "未知部门",
				})
			}
		} else {
			deptInfoList = append(deptInfoList, map[string]string{
				"code": deptCode,
				"name": dept.DeptName,
			})
		}
	}
	return deptInfoList
}

// formatDeptInfo 格式化部门信息为字符串
func (l *ListMyDocumentTypeLogic) formatDeptInfo(deptInfoList []map[string]string) string {
	if len(deptInfoList) == 0 {
		return "无部门信息"
	}

	parts := make([]string, 0, len(deptInfoList))
	for _, deptInfo := range deptInfoList {
		parts = append(parts, fmt.Sprintf("%s(%s)", deptInfo["name"], deptInfo["code"]))
	}
	return strings.Join(parts, "/")
}

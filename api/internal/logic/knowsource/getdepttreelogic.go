// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

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

type GetDeptTreeLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 获取部门树
func NewGetDeptTreeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetDeptTreeLogic {
	return &GetDeptTreeLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetDeptTreeLogic) GetDeptTree(req *types.KnowsourceDeptTreeRequest) (resp *types.KnowsourceDeptTreeResponse, err error) {
	// 构建查询条件
	var conditions []string
	var args []interface{}

	clientId, _ := l.ctx.Value("clientId").(string)
	clientId = strings.TrimSpace(clientId)
	if clientId == "" {
		return &types.KnowsourceDeptTreeResponse{
			Response: types.Response{
				Code:    response.UnauthorizedCode,
				Message: "clientId不能为空，请重新登录",
			},
			Data: &types.KnowsourceDeptTreeData{
				Tree: []types.KnowsourceDeptTreeNode{},
			},
		}, nil
	}

	conditions = append(conditions, "client_id = ?")
	args = append(args, clientId)

	// 如果指定了 ParentCode，只作为根节点筛选条件，不限制查询范围
	// 我们需要查询所有相关节点来构建完整的树
	whereClause := ""
	if len(conditions) > 0 {
		whereClause = "WHERE " + strings.Join(conditions, " AND ")
	}

	// 查询所有部门数据（根据 GSDM 筛选，不限制 ParentCode）
	query := fmt.Sprintf(`
		SELECT 
			dept_code,
			dept_name,
			parent_code,
			gsdm,
			grade,
			end_mark,
			kind,
			b0110
		FROM fr_dept
		%s
		ORDER BY grade, dept_code
	`, whereClause)

	type DeptRow struct {
		DeptCode   string `db:"dept_code"`
		DeptName   string `db:"dept_name"`
		ParentCode string `db:"parent_code"`
		GSDM       string `db:"gsdm"`
		Grade      int64  `db:"grade"`
		EndMark    string `db:"end_mark"`
		Kind       string `db:"kind"`
		B0110      string `db:"b0110"`
	}

	var rows []DeptRow
	err = l.svcCtx.Mysql.QueryRowsCtx(l.ctx, &rows, query, args...)
	if err != nil {
		if err == sqlx.ErrNotFound {
			return &types.KnowsourceDeptTreeResponse{
				Response: types.Response{
					Code:    response.SuccessCode,
					Message: "查询成功",
				},
				Data: &types.KnowsourceDeptTreeData{
					Tree: []types.KnowsourceDeptTreeNode{},
				},
			}, nil
		}
		l.Logger.Errorf("获取部门数据失败: %v", err)
		return &types.KnowsourceDeptTreeResponse{
			Response: types.Response{
				Code:    response.ServerErrorCode,
				Message: "获取部门数据失败",
				Info:    err.Error(),
			},
		}, nil
	}

	if len(rows) == 0 {
		return &types.KnowsourceDeptTreeResponse{
			Response: types.Response{
				Code:    response.SuccessCode,
				Message: "查询成功",
			},
			Data: &types.KnowsourceDeptTreeData{
				Tree: []types.KnowsourceDeptTreeNode{},
			},
		}, nil
	}

	// 构建节点映射（使用指针，避免值拷贝问题）
	nodeMap := make(map[string]*types.KnowsourceDeptTreeNode)

	// 第一步：创建所有节点
	for _, row := range rows {
		node := &types.KnowsourceDeptTreeNode{
			DeptCode:   row.DeptCode,
			DeptName:   row.DeptName,
			ParentCode: row.ParentCode,
			GSDM:       row.GSDM,
			Grade:      row.Grade,
			EndMark:    row.EndMark,
			Kind:       row.Kind,
			B0110:      row.B0110,
			Children:   []types.KnowsourceDeptTreeNode{},
		}
		nodeMap[row.DeptCode] = node
	}

	// 第二步：递归构建树结构
	// 这个函数会为指定节点构建完整的子树（包括所有子节点）
	var buildSubTree func(deptCode string) types.KnowsourceDeptTreeNode
	buildSubTree = func(deptCode string) types.KnowsourceDeptTreeNode {
		node := nodeMap[deptCode]
		if node == nil {
			return types.KnowsourceDeptTreeNode{}
		}

		// 创建当前节点的副本
		result := types.KnowsourceDeptTreeNode{
			DeptCode:   node.DeptCode,
			DeptName:   node.DeptName,
			ParentCode: node.ParentCode,
			GSDM:       node.GSDM,
			Grade:      node.Grade,
			EndMark:    node.EndMark,
			Kind:       node.Kind,
			B0110:      node.B0110,
			Children:   []types.KnowsourceDeptTreeNode{},
		}

		// 查找所有子节点并递归构建
		for _, row := range rows {
			if row.ParentCode == deptCode {
				childTree := buildSubTree(row.DeptCode)
				result.Children = append(result.Children, childTree)
			}
		}

		return result
	}

	// 第三步：确定根节点并构建完整的树
	var rootNodes []types.KnowsourceDeptTreeNode
	for _, row := range rows {
		// 判断是否为根节点
		isRoot := (row.ParentCode == "" || nodeMap[row.ParentCode] == nil)

		if isRoot {
			// 如果指定了 ParentCode 作为根节点筛选条件，只添加匹配的根节点
			if req.ParentCode != "" {
				if row.DeptCode == req.ParentCode {
					rootNodes = append(rootNodes, buildSubTree(row.DeptCode))
				}
			} else {
				rootNodes = append(rootNodes, buildSubTree(row.DeptCode))
			}
		}
	}

	tree := rootNodes

	return &types.KnowsourceDeptTreeResponse{
		Response: types.Response{
			Code:    response.SuccessCode,
			Message: "查询成功",
		},
		Data: &types.KnowsourceDeptTreeData{
			Tree: tree,
		},
	}, nil
}

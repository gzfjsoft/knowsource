package model

import (
	"context"
	"fmt"
	"strings"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ FrDeptModel = (*customFrDeptModel)(nil)

type (
	// FrDeptModel is an interface to be customized, add more methods here,
	// and implement the added methods in customFrDeptModel.
	FrDeptModel interface {
		frDeptModel
		WithSession(session sqlx.Session) FrDeptModel

		// FindAllLiteByClientId 查询构建树所需字段
		FindAllLiteByClientId(ctx context.Context, clientId string) ([]*FrDeptLite, error)
		// HasChildren 判断某部门是否有子部门
		HasChildren(ctx context.Context, clientId, deptCode string) (bool, error)
		// DeleteByClientIdDeptCodes 按租户批量删除部门（dept_code in (...)）
		DeleteByClientIdDeptCodes(ctx context.Context, clientId string, deptCodes []string) error
		// UpdateParentAndGradeByClientIdDeptCode 更新父编码 + grade（按租户 + dept_code）
		UpdateParentAndGradeByClientIdDeptCode(ctx context.Context, clientId, deptCode, newParentCode string, newGrade int64) error
		// UpdateGradeByClientIdDeptCode 仅更新 grade（按租户 + dept_code）
		UpdateGradeByClientIdDeptCode(ctx context.Context, clientId, deptCode string, newGrade int64) error
	}

	customFrDeptModel struct {
		*defaultFrDeptModel
	}

	FrDeptLite struct {
		DeptCode   string `db:"dept_code"`
		ParentCode string `db:"parent_code"`
		Grade      int64  `db:"grade"`
	}
)

// NewFrDeptModel returns a model for the database table.
func NewFrDeptModel(conn sqlx.SqlConn) FrDeptModel {
	return &customFrDeptModel{
		defaultFrDeptModel: newFrDeptModel(conn),
	}
}

func (m *customFrDeptModel) WithSession(session sqlx.Session) FrDeptModel {
	return NewFrDeptModel(sqlx.NewSqlConnFromSession(session))
}

func (m *customFrDeptModel) FindAllLiteByClientId(ctx context.Context, clientId string) ([]*FrDeptLite, error) {
	query := fmt.Sprintf("SELECT `dept_code`,`parent_code`,`grade` FROM %s WHERE `client_id` = ?", m.tableName())
	var rows []*FrDeptLite
	err := m.conn.QueryRowsCtx(ctx, &rows, query, clientId)
	return rows, err
}

func (m *customFrDeptModel) HasChildren(ctx context.Context, clientId, deptCode string) (bool, error) {
	query := fmt.Sprintf("SELECT 1 FROM %s WHERE `client_id` = ? AND `parent_code` = ? LIMIT 1", m.tableName())
	var one int
	err := m.conn.QueryRowCtx(ctx, &one, query, clientId, deptCode)
	switch err {
	case nil:
		return true, nil
	case sqlx.ErrNotFound:
		return false, nil
	default:
		return false, err
	}
}

func (m *customFrDeptModel) DeleteByClientIdDeptCodes(ctx context.Context, clientId string, deptCodes []string) error {
	if len(deptCodes) == 0 {
		return nil
	}
	holders := strings.Repeat("?,", len(deptCodes))
	holders = strings.TrimSuffix(holders, ",")
	query := fmt.Sprintf("DELETE FROM %s WHERE `client_id` = ? AND `dept_code` IN (%s)", m.tableName(), holders)
	args := make([]interface{}, 0, 1+len(deptCodes))
	args = append(args, clientId)
	for _, c := range deptCodes {
		args = append(args, c)
	}
	_, err := m.conn.ExecCtx(ctx, query, args...)
	return err
}

func (m *customFrDeptModel) UpdateParentAndGradeByClientIdDeptCode(ctx context.Context, clientId, deptCode, newParentCode string, newGrade int64) error {
	query := fmt.Sprintf("UPDATE %s SET `parent_code` = ?, `grade` = ? WHERE `client_id` = ? AND `dept_code` = ?", m.tableName())
	_, err := m.conn.ExecCtx(ctx, query, newParentCode, newGrade, clientId, deptCode)
	return err
}

func (m *customFrDeptModel) UpdateGradeByClientIdDeptCode(ctx context.Context, clientId, deptCode string, newGrade int64) error {
	query := fmt.Sprintf("UPDATE %s SET `grade` = ? WHERE `client_id` = ? AND `dept_code` = ?", m.tableName())
	_, err := m.conn.ExecCtx(ctx, query, newGrade, clientId, deptCode)
	return err
}

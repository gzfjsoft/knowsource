package model

import (
	"context"
	"fmt"
	"strings"

	"knowsource/common/cryptx"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ EmpPasswordModel = (*customEmpPasswordModel)(nil)

type (
	// EmpPasswordModel is an interface to be customized, add more methods here,
	// and implement the added methods in customEmpPasswordModel.
	EmpPasswordModel interface {
		empPasswordModel
		WithSession(session sqlx.Session) EmpPasswordModel
		// InitWithPassword 使用与登录相同的 cryptx.PasswordEncrypt(salt, plainPassword) 写入密文
		InitWithPassword(ctx context.Context, salt string, plainPassword string) error
		// FindOneByClientIdEmpCode 按租户查询（gen 的 FindOne 仅按 emp_code，业务层应使用本方法）
		FindOneByClientIdEmpCode(ctx context.Context, clientId, empCode string) (*EmpPassword, error)
	}

	customEmpPasswordModel struct {
		*defaultEmpPasswordModel
	}
)

// NewEmpPasswordModel returns a model for the database table.
func NewEmpPasswordModel(conn sqlx.SqlConn) EmpPasswordModel {
	return &customEmpPasswordModel{
		defaultEmpPasswordModel: newEmpPasswordModel(conn),
	}
}

func (m *customEmpPasswordModel) WithSession(session sqlx.Session) EmpPasswordModel {
	return NewEmpPasswordModel(sqlx.NewSqlConnFromSession(session))
}

// FindOneByClientIdEmpCode 使用 client_id + emp_code 查询，不得调用 gen 的 FindOne（其仅支持 emp_code）
func (m *customEmpPasswordModel) FindOneByClientIdEmpCode(ctx context.Context, clientId, empCode string) (*EmpPassword, error) {
	query := fmt.Sprintf("select %s from %s where `client_id` = ? and `emp_code` = ? limit 1", empPasswordRows, m.tableName())
	var resp EmpPassword
	err := m.conn.QueryRowCtx(ctx, &resp, query, clientId, empCode)
	switch err {
	case nil:
		return &resp, nil
	case sqlx.ErrNotFound:
		return nil, ErrNotFound
	default:
		return nil, err
	}
}

// Update 按 tenant 更新：WHERE client_id + emp_code（覆盖 gen 里仅按 emp_code 的实现）
func (m *customEmpPasswordModel) Update(ctx context.Context, data *EmpPassword) error {
	query := fmt.Sprintf("update %s set `client_id` = ?, `password` = ? where `client_id` = ? and `emp_code` = ?", m.tableName())
	_, err := m.conn.ExecCtx(ctx, query, data.ClientId, data.Password, data.ClientId, data.EmpCode)
	return err
}

func (m *customEmpPasswordModel) InitWithPassword(ctx context.Context, salt string, plainPassword string) error {
	if strings.TrimSpace(salt) == "" {
		return fmt.Errorf("InitWithPassword: salt 不能为空")
	}
	if strings.TrimSpace(plainPassword) == "" {
		return fmt.Errorf("InitWithPassword: plainPassword 不能为空")
	}
	hash := cryptx.PasswordEncrypt(salt, plainPassword)
	sql := `INSERT INTO emp_password (client_id, emp_code, password)
SELECT e.client_id, e.femp_code, ?
FROM fr_emp e
WHERE NOT EXISTS (
  SELECT 1 FROM emp_password p WHERE p.client_id = e.client_id AND p.emp_code = e.femp_code
)`
	_, err := m.conn.ExecCtx(ctx, sql, hash)
	return err
}

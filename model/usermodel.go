package model

import (
	"context"
	"database/sql"
	"fmt"
	"knowsource/common/jwtx"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

// UserRankInfo 用户积分排名信息
type UserRankInfo struct {
	EmpId  int64  `json:"empId"`
	Name   string `json:"name"`
	Rank   int64  `json:"rank"`
	OrgId  int64  `json:"orgId"`
	Points int64  `json:"points"`
}

var _ UserModel = (*customUserModel)(nil)

type (
	// UserModel is an interface to be customized, add more methods here,
	// and implement the added methods in customUserModel.
	UserModel interface {
		userModel
		FindByPhone(ctx context.Context, phone string) (*User, error)
		FindByEmpId(ctx context.Context, empId string) (*User, error)
		FindById(ctx context.Context, id int64) (*User, error)
		UpdateAuditStatus(ctx context.Context, newData *User) error
		CountByOrgId(ctx context.Context, orgId int64) (int64, error)
		ExistPhoneOrEmpId(ctx context.Context, phone, empId string) (bool, error)
		FindByEmpIdAndPhone(ctx context.Context, empId, phone string) (*User, error)
		FindByEmpIdsAndPhones(ctx context.Context, empIds, phones []string, companyId int64) ([]*User, error)
		SaveRegister(ctx context.Context, user *User) error
		WithSession(session sqlx.Session) UserModel
		// FindPage finds users with pagination and optional filters
		// keyword is used for fuzzy search on name and exact match on emp_id
		FindPage(ctx context.Context, page, pageSize uint64, orgId, status int64, empId string, companyId int64, phone string, keyword string) ([]*User, int64, error)
		GenToken(ctx context.Context, user *User, accessExpire int64, accessSecret string, isAdmin int64, roleIds string) (string, error)
		UpdatePoints(ctx context.Context, userId, points, level int64) (int64, error)
		UpdateAvatar(ctx context.Context, userId int64, avatar string) error
		UpdateType(ctx context.Context, userId, typee int64) error
	}

	customUserModel struct {
		*defaultUserModel
	}
)

// NewUserModel returns a model for the database table.
func NewUserModel(conn sqlx.SqlConn) UserModel {
	return &customUserModel{
		defaultUserModel: newUserModel(conn),
	}
}

func (m *customUserModel) WithSession(session sqlx.Session) UserModel {
	return NewUserModel(sqlx.NewSqlConnFromSession(session))
}

func (m *customUserModel) FindByPhone(ctx context.Context, phone string) (*User, error) {
	var resp User
	query := fmt.Sprintf("select %s from %s where `phone` = ? AND is_deleted = 0 limit 1", userRows, m.table)
	err := m.conn.QueryRowCtx(ctx, &resp, query, phone)
	switch err {
	case nil:
		return &resp, nil
	case sqlx.ErrNotFound, sql.ErrNoRows:
		return nil, ErrNotFound
	default:
		return nil, err
	}
}

func (m *customUserModel) FindByEmpId(ctx context.Context, empId string) (*User, error) {
	var resp User
	query := fmt.Sprintf("SELECT %s FROM %s WHERE `emp_id` = ? AND is_deleted = 0 LIMIT 1", userRows, m.table)
	err := m.conn.QueryRowCtx(ctx, &resp, query, empId)
	switch err {
	case nil:
		return &resp, nil
	case sqlx.ErrNotFound, sql.ErrNoRows:
		return nil, ErrNotFound
	default:
		return nil, err
	}
}

func (m *customUserModel) UpdateAuditStatus(ctx context.Context, newData *User) error {
	query := fmt.Sprintf("update %s set `status` = ?, `audit_status` = ?, `audit_user_id` = ?, `audit_time` = ? where `id` = ?", m.table)
	_, err := m.conn.ExecCtx(ctx, query, newData.Status, newData.AuditStatus, newData.AuditUserId, newData.AuditTime, newData.Id)
	return err
}

func (m *customUserModel) CountByOrgId(ctx context.Context, orgId int64) (int64, error) {
	query := fmt.Sprintf("select count(*) from %s where `org_id` = ? AND is_deleted = 0", m.table)
	var count int64
	err := m.conn.QueryRowCtx(ctx, &count, query, orgId)
	return count, err
}

func (m *customUserModel) FindById(ctx context.Context, id int64) (*User, error) {
	var resp User
	query := fmt.Sprintf("select %s from %s where `id` = ? AND is_deleted = 0 limit 1", userRows, m.table)
	err := m.conn.QueryRowCtx(ctx, &resp, query, id)
	switch err {
	case nil:
		return &resp, nil
	case sqlx.ErrNotFound, sql.ErrNoRows:
		return nil, errors.New("找不到此账号")
	default:
		return nil, err
	}
}

// FindPage finds users with pagination and optional filters
func (m *customUserModel) FindPage(ctx context.Context, page, pageSize uint64, orgId, status int64, empId string, companyId int64, phone, keyword string) ([]*User, int64, error) {
	// Build where clause and args
	var whereClause strings.Builder
	args := make([]interface{}, 0)
	whereClause.WriteString("WHERE is_deleted = 0")

	if orgId > 0 {
		whereClause.WriteString(" AND org_id = ?")
		args = append(args, orgId)
	}

	if status > -1 {
		whereClause.WriteString(" AND audit_status = ?")
		args = append(args, status)
	}

	if companyId > 0 {
		whereClause.WriteString(" AND company_id = ?")
		args = append(args, companyId)
	}

	// Add keyword search condition
	if keyword != "" {
		whereClause.WriteString(" AND (name LIKE ?)")
		args = append(args, "%"+keyword+"%")
	}

	if empId != "" {
		whereClause.WriteString(" AND (emp_id = ?)")
		args = append(args, empId)
	}

	if phone != "" {
		whereClause.WriteString(" AND (phone = ?)")
		args = append(args, phone)
	}

	// Get total count
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM %s %s", m.table, whereClause.String())
	var total int64
	err := m.conn.QueryRowCtx(ctx, &total, countQuery, args...)
	if err != nil {
		return nil, 0, err
	}

	// Get paginated data
	query := fmt.Sprintf("SELECT %s FROM %s %s ORDER BY id ASC LIMIT ? OFFSET ?",
		userRows, m.table, whereClause.String())

	offset := (page - 1) * pageSize
	args = append(args, pageSize, offset)

	var users []*User
	err = m.conn.QueryRowsCtx(ctx, &users, query, args...)
	if err != nil {
		return nil, 0, err
	}

	return users, total, nil
}

func (m *customUserModel) GenToken(ctx context.Context, user *User, accessExpire int64, _ string, isAdmin int64, roleIds string) (string, error) {
	clientId, _ := ctx.Value("clientId").(string)
	token, err := jwtx.GenerateTokenWithContext(
		ctx,
		clientId,
		user.Id,
		user.EmpId,
		isAdmin,
		// user.CompanyId,
		user.Name,
		// user.OrgId,
		roleIds,
		time.Duration(accessExpire)*time.Second,
	)
	if err != nil {
		return "", err
	}

	// 更新用户登录信息
	user.LastLoginAt = time.Now()
	user.LoginCount++
	query := fmt.Sprintf("update %s set `last_login_at` = ?, `login_count` = ? where `id` = ?", m.table)
	_, err = m.conn.ExecCtx(ctx, query, user.LastLoginAt, user.LoginCount, user.Id)
	if err != nil {
		return "", err
	}

	return token, nil
}

func (m *customUserModel) FindByEmpIdAndPhone(ctx context.Context, empId, phone string) (*User, error) {
	var resp User
	query := fmt.Sprintf("select %s from %s where `emp_id` = ? and `phone` = ? and is_deleted = 0 limit 1", userRows, m.table)
	err := m.conn.QueryRowCtx(ctx, &resp, query, empId, phone)
	switch err {
	case nil:
		return &resp, nil
	case sqlx.ErrNotFound, sql.ErrNoRows:
		return nil, ErrNotFound
	default:
		return nil, err
	}
}

func (m *customUserModel) ExistPhoneOrEmpId(ctx context.Context, empId, phone string) (bool, error) {
	var count int64
	query := fmt.Sprintf("select count(0) from %s where (`emp_id` = ? or `phone` = ?) and is_deleted = 0 limit 1", m.table)
	err := m.conn.QueryRowCtx(ctx, &count, query, empId, phone)
	return count > 0, err
}

func (m *customUserModel) FindByEmpIdsAndPhones(ctx context.Context, empIds, phones []string, companyId int64) ([]*User, error) {
	empIdsJoin := strings.Join(empIds, "','")
	phonesJoin := strings.Join(phones, "','")
	var resp []*User
	query := fmt.Sprintf("select %s from %s where (`emp_id` IN (%s) or `phone` IN (%s)) and is_deleted = 0 and company_id = ? ", userRows, m.table, "'"+empIdsJoin+"'", "'"+phonesJoin+"'")
	err := m.conn.QueryRowsCtx(ctx, &resp, query, companyId)
	switch err {
	case nil:
		return resp, nil
	case sqlx.ErrNotFound, sql.ErrNoRows:
		return nil, ErrNotFound
	default:
		return nil, err
	}
}

func (m *customUserModel) SaveRegister(ctx context.Context, user *User) error {
	query := fmt.Sprintf("update %s set `audit_status` = ? where `id` = ?", m.table)
	_, err := m.conn.ExecCtx(ctx, query, user.AuditStatus, user.Id)
	return err
}

func (m *customUserModel) UpdatePoints(ctx context.Context, userId, points, level int64) (int64, error) {
	return 0, nil
	// canUpgrade, newLevel := CheckIsNewUserPointLevel(points, level)
	// if canUpgrade {
	// 	level = newLevel
	// }
	// query := fmt.Sprintf("update %s set `points` = ? , points_level = ? where `id` = ?", m.table)
	// _, err := m.conn.ExecCtx(ctx, query, points, level, userId)
	// return level, err
}

func (m *customUserModel) UpdateType(ctx context.Context, userId, typee int64) error {
	query := fmt.Sprintf("update %s set `type` = %d where `id` = %d", m.table, typee, userId)
	_, err := m.conn.ExecCtx(ctx, query)
	return err
}

func (m *customUserModel) UpdateAvatar(ctx context.Context, userId int64, avatar string) error {
	query := fmt.Sprintf("update %s set `avatar` = '%s' where `id` = %d", m.table, avatar, userId)
	_, err := m.conn.ExecCtx(ctx, query)
	return err
}

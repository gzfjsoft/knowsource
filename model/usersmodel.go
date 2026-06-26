package model

import (
	"context"
	"errors"
	"fmt"
	"knowsource/common/cryptx"
	"knowsource/consts"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ UsersModel = (*customUsersModel)(nil)

type (
	// UsersModel is an interface to be customized, add more methods here,
	// and implement the added methods in customUsersModel.
	UsersModel interface {
		usersModel
		WithSession(session sqlx.Session) UsersModel
		RestPass(ctx context.Context, data *Users) error
		FindOneByPhone(ctx context.Context, phone string) (*Users, error)
		FindOneByUsername(ctx context.Context, username string) (*Users, error)

		FindUsers(ctx context.Context, username, email, phone string, page, pageSize uint64) ([]*Users, uint64, error)

		FindUsersNames(ctx context.Context, userids string) ([]*UserName, error)

		UpdateInfo(ctx context.Context, newData *Users) error
		FindOneByX(ctx context.Context, AccX string) (*Users, error)
		FindOneWithDelete(ctx context.Context, userId uint64) (*Users, error)
		DeleteWithDelete(ctx context.Context, userId uint64) error
		FindOneByGoogle(ctx context.Context, AccGoogle string) (*Users, error)
		FindByOpenId(ctx context.Context, OpenId string) (*Users, error)
		FindByUnionId(ctx context.Context, UnionId string) (*Users, error)
		Count(ctx context.Context, condition string) (uint64, error)
		RegisterNewUserinDB(ctx context.Context, Username, Nickname, Email, Phone, Password, HeadUrl, OpenId, UnionId, AccGoogle, AccX, passsalt string) (uint64, error)
		FindOrCreateUserByUnionID(openid string, unionid string, nickname string, headImgURL string, salt string) (*Users, error, bool)
		FindOneByEmail(ctx context.Context, email string) (*Users, error)
		FindByUserIds(ctx context.Context, userIds []uint64) ([]*Users, error)
	}

	customUsersModel struct {
		*defaultUsersModel
	}
)

// NewUsersModel returns a model for the database table.
func NewUsersModel(conn sqlx.SqlConn) UsersModel {
	return &customUsersModel{
		defaultUsersModel: newUsersModel(conn),
	}
}

func (m *customUsersModel) WithSession(session sqlx.Session) UsersModel {
	return NewUsersModel(sqlx.NewSqlConnFromSession(session))
}

//

// func FindOrCreateUserByUnionID(svcCtx *svc.ServiceContext, openid string, nickname string) (*model.Users, error) {
// 	ctx := context.Background()

// 	user, err := svcCtx.UsersModel.FindByUnionId(ctx, openid)

// 	if err == sqlx.ErrNotFound {
// 		//生成随机用户名 32位
// 		username := uuid.New().String()[:32]

// 		userId, err := svcCtx.UsersModel.RegisterNewUserinDB(context.Background(),
// 			username,
// 			nickname,
// 			openid+"@weixin.qq.com",
// 			"",
// 			"",
// 			"",
// 			openid,
// 			"",
// 			"",
// 			svcCtx.Config.Salt,
// 		)

// 		if err != nil {
// 			return nil, err
// 		}

// 		user, err = svcCtx.UsersModel.FindOne(ctx, userId)
// 		if err != nil {
// 			return nil, err
// 		}

// 		return user, nil

// 	} else if err != nil {
// 		logx.Error("Failed to find user", err)
// 		return nil, err
// 	}
// 	return user, nil

// }

// func (l *WechatOfficalCallbackLogic) FindOrCreateUserByUnionID(svcCtx *svc.ServiceContext, openid string, nickname string, headImgURL string) (*model.Users, error,bool) {
// 	ctx := context.Background()

// 	user, err := svcCtx.UsersModel.FindByUnionId(ctx, openid)

// 	if err == sqlx.ErrNotFound {
// 		//生成随机用户名 32位
// 		username := uuid.New().String()[:32]

// 		userId, err := l.svcCtx.UsersModel.RegisterNewUserinDB(l.ctx,
// 			username,
// 			nickname,
// 			openid+"@weixin.qq.com",
// 			"",
// 			"",
// 			headImgURL,
// 			openid,
// 			"",
// 			"",
// 			l.svcCtx.Config.Salt,
// 		)

// 		if err != nil {
// 			return nil, err
// 		}

// 		user, err = l.svcCtx.UsersModel.FindOne(ctx, userId)
// 		if err != nil {
// 			return nil, err
// 		}

// 		return user, nil

// 	} else if err != nil {
// 		logx.Error("Failed to find user", err)
// 		return nil, err
// 	}
// 	return user, nil

// }

func (m *customUsersModel) FindOrCreateUserByUnionID(openid string, unionid string, nickname string, headImgURL string, salt string) (*Users, error, bool) {
	ctx := context.Background()

	isNewUser := false
	var user *Users

	var err error
	if unionid != "" {
		user, err = m.FindByUnionId(ctx, unionid)
	} else {
		logx.Errorf("unionid is empty, use openid to find user")
		user, err = m.FindByOpenId(ctx, openid)
	}

	if err == sqlx.ErrNotFound {
		isNewUser = true
		//生成随机用户名 32位
		username := uuid.New().String()[:32]

		userId, err := m.RegisterNewUserinDB(context.Background(),
			username,
			nickname,
			openid+"@weixin.qq.com",
			"",
			"",
			headImgURL,
			openid,
			unionid,
			"",
			"",
			salt,
		)

		if err != nil {
			return nil, err, false
		}

		user, err = m.FindOne(ctx, userId)
		if err != nil {
			return nil, err, false
		}

		return user, nil, true
	} else if err != nil {
		logx.Error("Failed to find user", err)
		return nil, err, false
	}
	return user, nil, isNewUser

}

func (m *customUsersModel) DeleteWithDelete(ctx context.Context, userId uint64) error {
	query := fmt.Sprintf("delete from %s where `user_id` = ?", m.table)
	_, err := m.conn.ExecCtx(ctx, query, userId)
	return err
}

type UserName struct {
	UserId   uint64 `db:"user_id"`
	Username string `db:"username"`
}

func (m *customUsersModel) FindUsersNames(ctx context.Context, userids string) ([]*UserName, error) {

	query := fmt.Sprintf("select user_id,username from %s where `user_id` in (%s)", m.table, userids)
	var resp []*UserName
	err := m.conn.QueryRowsCtx(ctx, &resp, query)
	return resp, err
}

func (m *customUsersModel) FindByOpenId(ctx context.Context, OpenId string) (*Users, error) {
	//TODO
	var resp Users
	query := fmt.Sprintf("select %s from %s where `open_id` = ? and is_deleted=0 limit 1", usersRows, m.table)
	err := m.conn.QueryRowCtx(ctx, &resp, query, OpenId)
	switch err {
	case nil:
		return &resp, nil
	case sqlx.ErrNotFound:
		return nil, ErrNotFound
	default:
		return nil, err
	}
}

func (m *customUsersModel) FindByUnionId(ctx context.Context, UnionId string) (*Users, error) {
	//TODO
	var resp Users
	query := fmt.Sprintf("select %s from %s where `union_id` = ? and is_deleted=0 limit 1", usersRows, m.table)
	err := m.conn.QueryRowCtx(ctx, &resp, query, UnionId)
	switch err {
	case nil:
		return &resp, nil
	case sqlx.ErrNotFound:
		return nil, ErrNotFound
	default:
		return nil, err
	}
}
func (m *customUsersModel) FindOneByX(ctx context.Context, AccX string) (*Users, error) {
	var resp Users
	query := fmt.Sprintf("select %s from %s where `acc_x` = ? and is_deleted=0 limit 1", usersRows, m.table)
	err := m.conn.QueryRowCtx(ctx, &resp, query, AccX)
	switch err {
	case nil:
		return &resp, nil
	case sqlx.ErrNotFound:
		return nil, ErrNotFound
	default:
		return nil, err
	}
}

func (m *customUsersModel) FindOneByGoogle(ctx context.Context, AccGoogle string) (*Users, error) {
	var resp Users
	query := fmt.Sprintf("select %s from %s where `acc_google` = ? and is_deleted=0 limit 1", usersRows, m.table)
	err := m.conn.QueryRowCtx(ctx, &resp, query, AccGoogle)
	switch err {
	case nil:
		return &resp, nil
	case sqlx.ErrNotFound:
		return nil, ErrNotFound
	default:
		return nil, err
	}
}

func (m *customUsersModel) FindOneByUsername(ctx context.Context, username string) (*Users, error) {
	var resp Users
	query := fmt.Sprintf("select %s from %s where `username` = ? and is_deleted=0 limit 1", usersRows, m.table)
	err := m.conn.QueryRowCtx(ctx, &resp, query, username)
	return &resp, err
}

func (m *customUsersModel) FindOneByEmail(ctx context.Context, email string) (*Users, error) {
	var resp Users
	query := fmt.Sprintf("select %s from %s where `email` = ? and is_deleted=0 limit 1", usersRows, m.table)
	err := m.conn.QueryRowCtx(ctx, &resp, query, email)
	switch err {
	case nil:
		return &resp, nil
	case sqlx.ErrNotFound:
		return nil, ErrNotFound
	default:
		return nil, err
	}
}

func (m *customUsersModel) UpdateInfo(ctx context.Context, newData *Users) error {

	usersRowsWithPlaceHolder2 := strings.Join([]string{"`username`", "`nickname`", "`email`", "`phone`"}, "=?,") + "=?"

	logx.Infof("update %s set %s where `user_id` = ?", m.table, usersRowsWithPlaceHolder2)

	query := fmt.Sprintf("update %s set %s where `user_id` = ?", m.table, usersRowsWithPlaceHolder2)
	_, err := m.conn.ExecCtx(ctx, query, newData.Username, newData.Nickname, newData.Email, newData.Phone, newData.UserId)
	return err
}

func (m *customUsersModel) FindOneByPhone(ctx context.Context, phone string) (*Users, error) {
	var resp Users
	query := fmt.Sprintf("select %s from %s where `phone` = ? and is_deleted=0 limit 1", usersRows, m.table)
	err := m.conn.QueryRowCtx(ctx, &resp, query, phone)
	switch err {
	case nil:
		return &resp, nil
	case sqlx.ErrNotFound:
		return nil, ErrNotFound
	default:
		return nil, err
	}
}

func (m *customUsersModel) Delete(ctx context.Context, userId uint64) error {
	query := fmt.Sprintf("update %s set is_deleted=1 where `user_id` = ?", m.table)
	_, err := m.conn.ExecCtx(ctx, query, userId)
	return err
}

func (m *customUsersModel) FindOneWithDelete(ctx context.Context, userId uint64) (*Users, error) {
	query := fmt.Sprintf("select %s from %s where `user_id` = ?   limit 1", usersRows, m.table) //and
	var resp Users
	err := m.conn.QueryRowCtx(ctx, &resp, query, userId)
	switch err {
	case nil:
		return &resp, nil
	case sqlx.ErrNotFound:
		return nil, ErrNotFound
	default:
		return nil, err
	}
}

func (m *customUsersModel) FindOne(ctx context.Context, userId uint64) (*Users, error) {
	query := fmt.Sprintf("select %s from %s where `user_id` = ? and is_deleted=0 limit 1", usersRows, m.table) //and
	var resp Users
	err := m.conn.QueryRowCtx(ctx, &resp, query, userId)
	switch err {
	case nil:
		return &resp, nil
	case sqlx.ErrNotFound:
		return nil, ErrNotFound
	default:
		return nil, err
	}
}

func (m *customUsersModel) FindOneBySysRole(ctx context.Context, sysRole string) (*Users, error) {
	query := fmt.Sprintf("select %s from %s where `sys_role` = ?  limit 1", usersRows, m.table) //and
	var resp Users
	err := m.conn.QueryRowCtx(ctx, &resp, query, sysRole)
	switch err {
	case nil:
		return &resp, nil
	case sqlx.ErrNotFound:
		return nil, ErrNotFound
	default:
		return nil, err
	}
}

func (m *customUsersModel) RestPass(ctx context.Context, newData *Users) error {
	query := fmt.Sprintf("update %s set password_hash=? where `user_id` = ?", m.table)
	_, err := m.conn.ExecCtx(ctx, query, newData.PasswordHash, newData.UserId)
	return err
}

func (m *customUsersModel) FindUsers(ctx context.Context, username, email, phone string, page, pageSize uint64) ([]*Users, uint64, error) {
	//whereClause := "WHERE is_deleted = 0"
	whereClause := "WHERE 1 = 1"
	params := []interface{}{}

	if username != "" {
		whereClause += " AND username LIKE ?"
		params = append(params, fmt.Sprintf("%%%s%%", username))
	}
	if email != "" {
		whereClause += " AND email LIKE ?"
		params = append(params, fmt.Sprintf("%%%s%%", email))
	}
	if phone != "" {
		whereClause += " AND phone LIKE ?"
		params = append(params, fmt.Sprintf("%%%s%%", phone))
	}

	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM %s %s", m.table, whereClause)
	var total uint64
	err := m.conn.QueryRowCtx(ctx, &total, countQuery, params...)
	if err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	query := fmt.Sprintf("SELECT %s FROM %s %s LIMIT ? OFFSET ?", usersRows, m.table, whereClause)
	params = append(params, pageSize, offset)

	var users []*Users
	err = m.conn.QueryRowsCtx(ctx, &users, query, params...)
	if err != nil {
		return nil, 0, err
	}

	return users, total, nil
}

func (m *customUsersModel) Count(ctx context.Context, condition string) (uint64, error) {
	query := fmt.Sprintf("select count(*) from %s where %s", m.table, condition)
	var count uint64
	err := m.conn.QueryRowCtx(ctx, &count, query)
	return count, err
}

func (m *customUsersModel) RegisterNewUserinDB(ctx context.Context, Username, Nickname, Email, Phone, Password, HeadUrl, OpenId, UnionId, AccGoogle, AccX, passsalt string) (uint64, error) {

	PasswordHash := cryptx.PasswordEncrypt(passsalt, Password)
	if Password == "" {
		PasswordHash = ""
	}

	var userId int64
	var orgid uint64

	//get default	 orgid
	OrganizationModel := NewOrganizationsModel(m.conn)
	org, err := OrganizationModel.FindDefaultOne(ctx)
	if err != nil {
		logx.Infof("RegisterNewUserinDB.FindDefaultOne %v", err)
		return 0, err
	}
	orgid = org.OrgId

	//get sysrole superadmin

	superadmin, err := m.FindOneBySysRole(ctx, consts.SUPER_ADMIN)
	if err != nil {
		logx.Infof("RegisterNewUserinDB.FindOneBySysRole %v", err)
		return 0, err
	}
	if superadmin == nil {
		logx.Infof("RegisterNewUserinDB.superadmin is nil")
		return 0, errors.New("superadmin is nil")
	}

	//get supperadmin role
	rolesModel := NewRolesModel(m.conn)
	role, err := rolesModel.FindOneByName(ctx, consts.SUPER_ADMIN)
	if err != nil {
		logx.Infof("RegisterNewUserinDB.FindOneByName %v", err)
		return 0, err
	}
	if role == nil {
		logx.Infof("RegisterNewUserinDB.role is nil")
		return 0, errors.New("role is nil")
	}

	// 创建一个user

	err = m.conn.Transact(func(session sqlx.Session) error {
		// Create user with transaction
		usersModel := m.WithSession(session)
		uuid := uuid.New().String()
		res, err := usersModel.Insert(ctx, &Users{
			Username:     Username,
			Email:        Email,
			Phone:        Phone,
			Nickname:     Nickname,
			IsMaster:     1,
			ShareBalance: 1,
			HeadUrl:      HeadUrl,
			Avatar:       HeadUrl,
			Photo:        HeadUrl,
			Photos:       HeadUrl,
			PasswordHash: PasswordHash,
			AccX:         AccX,
			AccGoogle:    AccGoogle,
			OpenId:       OpenId,
			UnionId:      UnionId,
			SysRole:      "user",
			Uuid:         uuid,
			LoginedAt:    time.Now(),
			CreatedAt:    time.Now(),
			VipExpiredAt: time.Date(2000, 1, 1, 0, 0, 0, 0, time.Local),
		})

		if err != nil {
			return err
		}

		userId, err = res.LastInsertId()
		if err != nil {
			return err
		}

		logx.Infof("--> FindDefaultOne")

		// Link user to org with transaction
		newOrgsUsersModel := NewOrgsUsersModel(m.conn)
		orgsUsersModel := newOrgsUsersModel.WithSession(session)
		_, err = orgsUsersModel.Insert(ctx, &OrgsUsers{
			OrgId:  orgid,
			UserId: uint64(userId),
			Role:   "member", // Default role
		})

		if err != nil {
			logx.Infof("--> Insert OrgsUsers %v", err)
			return err
		}

		// TODO: magic number
		logx.Infof("<-- Insert userRolesModel")
		newUserRolesModel := NewUserRolesModel(m.conn)
		userRolesModel := newUserRolesModel.WithSession(session)
		_, err = userRolesModel.Insert(ctx, &UserRoles{
			UserId:     uint64(userId),
			RoleId:     role.RoleId,
			AssignedBy: superadmin.UserId,
			AssignedAt: time.Now(),
		})

		if err != nil {
			logx.Infof("--> Insert UserRoles %v", err)
			return err
		}

		logx.Infof("<-- Insert Balances")
		newBalancesModel := NewBalancesModel(m.conn)
		balanceModel := newBalancesModel.WithSession(session)

		item, err := balanceModel.FindOneByUserAndCurrency(ctx, uint64(userId), 1, "CNY")
		if err != nil && err != sqlx.ErrNotFound {
			logx.Infof("--> FindOneByUserAndCurrency %v", err)
			return err
		}

		logx.Infof("<-- Insert Balances")
		if err == sqlx.ErrNotFound {
			_, err = balanceModel.Insert(ctx, &Balances{
				UserId:       uint64(userId),
				OrgId:        orgid,
				CurrencyCode: "CNY",
				Balance:      0,
			})
			if err != nil {
				logx.Infof("--> Insert Balances %v", err)
				return err
			}
		} else {
			item.Balance = 0
			err = balanceModel.Update(ctx, item)
			if err != nil {
				logx.Infof("--> Update Balances %v", err)
				return err
			}
		}

		logx.Infof("<-- Insert Balances")
		return nil
	})

	if err != nil {
		return 0, err
	}

	return uint64(userId), nil

}

func (m *customUsersModel) FindByUserIds(ctx context.Context, userIds []uint64) ([]*Users, error) {
	var resp []*Users
	var userIdsStr string
	for _, userId := range userIds {
		userIdsStr += fmt.Sprintf("%d,", userId)
	}
	userIdsStr = strings.TrimSuffix(userIdsStr, ",")
	logx.Infof("userIdsStr: %s", userIdsStr)
	query := fmt.Sprintf("select %s from %s where `user_id` in (%s)", usersRows, m.table, userIdsStr)
	err := m.conn.QueryRowsCtx(ctx, &resp, query)
	return resp, err
}

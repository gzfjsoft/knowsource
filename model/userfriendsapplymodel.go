package model

import (
	"context"
	"fmt"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ UserFriendsApplyModel = (*customUserFriendsApplyModel)(nil)

type (
	// UserFriendsApplyModel is an interface to be customized, add more methods here,
	// and implement the added methods in customUserFriendsApplyModel.
	UserFriendsApplyModel interface {
		userFriendsApplyModel
		withSession(session sqlx.Session) UserFriendsApplyModel
		DeleteByUserIdFriendId(ctx context.Context, userId uint64, friendId uint64) error
		FindViewByContion(ctx context.Context, userid uint64) ([]*FriendsInfo, error)
		FindRecommands(whereBuilder string, pageSize, offset int64, args ...interface{}) ([]*FriendsInfo, error)
		FindVisitors(whereBuilder string, pageSize, offset int64, args ...interface{}) ([]*VisitorInfo, error)
		FindVisitorsToMe(whereBuilder string, pageSize, offset int64, args ...interface{}) ([]*VisitorInfo, error)
		FindFriends(whereBuilder string, pageSize, offset int64, args ...interface{}) ([]*FriendsExInfo, error)
		FindFriendsApply(whereBuilder string, pageSize, offset int64, args ...interface{}) ([]*FriendsInfo, error)
		FindFriendsApplyToMe(whereBuilder string, pageSize, offset int64, args ...interface{}) ([]*FriendsInfo, error)
	}

	customUserFriendsApplyModel struct {
		*defaultUserFriendsApplyModel
	}
)

// NewUserFriendsApplyModel returns a model for the database table.
func NewUserFriendsApplyModel(conn sqlx.SqlConn) UserFriendsApplyModel {
	return &customUserFriendsApplyModel{
		defaultUserFriendsApplyModel: newUserFriendsApplyModel(conn),
	}
}

func (m *customUserFriendsApplyModel) withSession(session sqlx.Session) UserFriendsApplyModel {
	return NewUserFriendsApplyModel(sqlx.NewSqlConnFromSession(session))
}

type VisitorInfo struct {
	FriendsInfo
	Times uint64 `db:"times"`
}
type FriendsExInfo struct {
	FriendsInfo
	ComeTimes uint64 `db:"come_times"`
	Times     uint64 `db:"times"`
}

type FriendsInfo struct {
	Id              uint64 `db:"id"`
	UserId          uint64 `db:"user_id"`
	FriendId        uint64 `db:"friend_id"`
	Uuid            string `db:"uuid"`
	Checked         int64  `db:"checked"`
	State           int64  `db:"state"`
	IsPhoneVerified uint64 `db:"is_phone_verified"`
	IsEmailVerified uint64 `db:"is_email_verified"`
	Nickname        string `db:"nickname"`
	Birthday        uint64 `db:"birthday"`
	Age             uint64 `db:"age"`
	Country         string `db:"country"`
	Province        string `db:"province"`
	City            string `db:"city"`
	District        string `db:"district"`
	Height          uint64 `db:"height"`
	Weight          uint64 `db:"weight"`
	Education       string `db:"education"`
	Sex             uint64 `db:"sex"`
	Occupation      string `db:"occupation"`
	Brief           string `db:"brief"`
	Income          uint64 `db:"income"`
	BoyCount        uint64 `db:"boy_count"`
	GirlCount       uint64 `db:"girl_count"`
	MaritalStatus   string `db:"marital_status"`
	HomeTown        string `db:"home_town"`
	HomeCity        string `db:"home_city"`
	HomeCountry     string `db:"home_country"`
	HomeProvince    string `db:"home_province"`
	HomeDistrict    string `db:"home_district"`
	Location        string `db:"location"`
	Religion        string `db:"religion"`
	Hobbies         string `db:"hobbies"`
	Tags            string `db:"tags"`
	Photo           string `db:"photo"`
	Photos          string `db:"photos"`
	Introduction    string `db:"introduction"`
	VipLevel        uint64 `db:"vip_level"`
	IsVerified      uint64 `db:"is_verified"`
}

func (m *customUserFriendsApplyModel) FindViewByContion(ctx context.Context, userid uint64) ([]*FriendsInfo, error) {
	sql := "SELECT a.id as id ,a.user_id as uid,b.*  FROM (user_friends_apply  a JOIN  users b ON (a.friend_id = b.user_id))"
	query := fmt.Sprintf(sql + " where a.user_id  = ? order by a.created_at asc")
	var replies []*FriendsInfo
	err := m.conn.QueryRowsCtx(ctx, &replies, query, userid)
	return replies, err
}

func (m *customUserFriendsApplyModel) DeleteByUserIdFriendId(ctx context.Context, userId uint64, friendId uint64) error {
	query := "DELETE FROM user_friends_apply WHERE user_id = ? AND friend_id = ?"
	_, err := m.conn.ExecCtx(ctx, query, userId, friendId)
	return err
}

func (m *customUserFriendsApplyModel) FindFriendsApply(whereBuilder string, pageSize, offset int64, args ...interface{}) ([]*FriendsInfo, error) {
	sql := "SELECT a.id as id ,a.user_id as uid,b.*  FROM (user_friends_apply  a JOIN  users b ON (a.friend_id = b.user_id))"

	query := fmt.Sprintf(sql+" %s ORDER BY created_at DESC LIMIT ? OFFSET ?", whereBuilder)
	args = append(args, pageSize, offset)

	var friends []*FriendsInfo
	err := m.conn.QueryRows(&friends, query, args...)
	return friends, err
}

func (m *customUserFriendsApplyModel) FindFriendsApplyToMe(whereBuilder string, pageSize, offset int64, args ...interface{}) ([]*FriendsInfo, error) {
	sql := "SELECT a.id as id ,a.user_id as uid,b.*  FROM (user_friends_apply  a JOIN  users b ON (a.user_id = b.user_id))"

	query := fmt.Sprintf(sql+" %s ORDER BY created_at DESC LIMIT ? OFFSET ?", whereBuilder)
	args = append(args, pageSize, offset)

	var friends []*FriendsInfo
	err := m.conn.QueryRows(&friends, query, args...)
	return friends, err
}

func (m *customUserFriendsApplyModel) FindVisitors(whereBuilder string, pageSize, offset int64, args ...interface{}) ([]*VisitorInfo, error) {
	sql := "SELECT a.id as id ,a.user_id as uid,a.times ,b.*  FROM (user_visitors  a JOIN  users b ON (a.friend_id = b.user_id))"

	if whereBuilder == "" {
		whereBuilder = " where a.times > 0"
	} else {
		whereBuilder += " and a.times > 0"
	}

	query := fmt.Sprintf(sql+" %s ORDER BY a.updated_at DESC LIMIT ? OFFSET ?", whereBuilder)
	args = append(args, pageSize, offset)

	var friends []*VisitorInfo
	err := m.conn.QueryRows(&friends, query, args...)
	return friends, err
}

func (m *customUserFriendsApplyModel) FindVisitorsToMe(whereBuilder string, pageSize, offset int64, args ...interface{}) ([]*VisitorInfo, error) {
	sql := "SELECT a.id as id ,a.user_id as uid,a.times, b.*  FROM (user_visitors  a JOIN  users b ON (a.user_id = b.user_id))"

	if whereBuilder == "" {
		whereBuilder = " where a.times > 0"
	} else {
		whereBuilder += " and a.times > 0"
	}

	query := fmt.Sprintf(sql+" %s ORDER BY  created_at DESC LIMIT ? OFFSET ?", whereBuilder)
	args = append(args, pageSize, offset)

	var friends []*VisitorInfo
	err := m.conn.QueryRows(&friends, query, args...)
	return friends, err
}
func (m *customUserFriendsApplyModel) FindFriends(whereBuilder string, pageSize, offset int64, args ...interface{}) ([]*FriendsExInfo, error) {
	sql := "SELECT a.id as id ,a.user_id as uid,b.*  FROM (user_friends  a JOIN  users b ON (a.friend_id = b.user_id))"

	query := fmt.Sprintf(sql+" %s ORDER BY created_at DESC LIMIT ? OFFSET ?", whereBuilder)
	args = append(args, pageSize, offset)

	var friends []*FriendsExInfo
	err := m.conn.QueryRows(&friends, query, args...)
	return friends, err
}

func (m *customUserFriendsApplyModel) FindRecommands(whereBuilder string, pageSize, offset int64, args ...interface{}) ([]*FriendsInfo, error) {
	sql := "SELECT a.id as id ,a.user_id as uid,b.*  FROM (user_recommends  a JOIN  users b ON (a.friend_id = b.user_id))"

	query := fmt.Sprintf(sql+" %s ORDER BY created_at DESC LIMIT ? OFFSET ?", whereBuilder)
	args = append(args, pageSize, offset)

	var friends []*FriendsInfo
	err := m.conn.QueryRows(&friends, query, args...)
	return friends, err
}

// SELECT a.id as id ,a.user_id as uid,b.*  FROM (user_recommends  a JOIN  users b ON (a.friend_id = b.user_id));
// SELECT a.id as id ,a.user_id as uid,b.*  FROM (user_visitors  a JOIN  users b ON (a.friend_id = b.user_id));
// SELECT a.id as id ,a.user_id as uid,b.*  FROM (user_friends  a JOIN  users b ON (a.friend_id = b.user_id));
// SELECT a.id as id ,a.user_id as uid,b.*  FROM (user_friends_apply  a JOIN  users b ON (a.friend_id = b.user_id)

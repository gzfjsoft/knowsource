package model

import (
	"context"
	"database/sql"
	"time"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ UserVisitorsModel = (*customUserVisitorsModel)(nil)

type (
	// UserVisitorsModel is an interface to be customized, add more methods here,
	// and implement the added methods in customUserVisitorsModel.
	UserVisitorsModel interface {
		userVisitorsModel
		withSession(session sqlx.Session) UserVisitorsModel
		AddVisited(userId, friendId uint64) error
	}

	customUserVisitorsModel struct {
		*defaultUserVisitorsModel
	}
)

// NewUserVisitorsModel returns a model for the database table.
func NewUserVisitorsModel(conn sqlx.SqlConn) UserVisitorsModel {
	return &customUserVisitorsModel{
		defaultUserVisitorsModel: newUserVisitorsModel(conn),
	}
}

func (m *customUserVisitorsModel) withSession(session sqlx.Session) UserVisitorsModel {
	return NewUserVisitorsModel(sqlx.NewSqlConnFromSession(session))
}

func (m *customUserVisitorsModel) AddVisited(userId, friendId uint64) error {
	ctx := context.Background()

	//找到我访问的记录
	v, err := m.FindOneByUserIdFriendId(ctx, userId, friendId)
	if err != nil && err != sql.ErrNoRows {
		return err
	}
	if err == sql.ErrNoRows {
		//insert
		v = &UserVisitors{
			UserId:    userId,
			FriendId:  friendId,
			Times:     0,
			ComeTimes: 0,
			VisiteAt:  time.Now().Add(-time.Minute * 10),
		}

		result, err := m.Insert(ctx, v)
		if err != nil {
			logx.Error("插入用户访问记录失败", err)
			return err
		}
		id, err := result.LastInsertId()
		if err != nil {
			logx.Error("插入用户访问记录失败LastInsertId", err)
			return err
		}
		v.Id = uint64(id)

	}
	//找到对方访问我的记录
	v2, err := m.FindOneByUserIdFriendId(ctx, friendId, userId)
	if err != nil && err != sql.ErrNoRows {
		return err
	}
	if err == sql.ErrNoRows {
		//insert
		v2 = &UserVisitors{
			UserId:    friendId,
			FriendId:  userId,
			Times:     0,
			ComeTimes: 0,
			VisiteAt:  time.Now().Add(-time.Minute * 10),
		}

		result, err := m.Insert(ctx, v2)
		if err != nil {
			logx.Error("插入用户访问记录失败", err)
			return err
		}
		id, err := result.LastInsertId()
		if err != nil {
			logx.Error("插入用户访问记录失败LastInsertId", err)
			return err
		}
		v2.Id = uint64(id)
	}

	now := time.Now()

	// 如果更新时间大于5分钟，则更新次数，否则不更新，5分钟才认为用户访问了

	if v.VisiteAt.Before(now.Add(-time.Minute * 2)) {
		v.Times++
		v.VisiteAt = now
		err = m.Update(ctx, v)
		if err != nil {
			logx.Error("更新用户访问记录失败", err)
		}

		v2.ComeTimes++
		err = m.Update(ctx, v2)
		if err != nil {
			logx.Error("更新用户访问记录失败", err)
		}
	}

	return err
}

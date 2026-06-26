package model

import (
	"context"
	"fmt"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ MomentModel = (*customMomentModel)(nil)

type (
	// MomentModel is an interface to be customized, add more methods here,
	// and implement the added methods in customMomentModel.
	MomentModel interface {
		momentModel
		withSession(session sqlx.Session) MomentModel
		FindAll(ctx context.Context) ([]*Moment, error)                         // Add this
		FindByUserId(ctx context.Context, userId int64) (*Moment, error)        // Add this
		FindByCondition(ctx context.Context, condition string) (*Moment, error) // Add this

		FindByConditions(ctx context.Context, condition string, values ...interface{}) (*[]Moment, error)
		Count(ctx context.Context, condition string, values ...interface{}) (int, error)
	}

	customMomentModel struct {
		*defaultMomentModel
	}
)

// NewMomentModel returns a model for the database table.
func NewMomentModel(conn sqlx.SqlConn) MomentModel {
	return &customMomentModel{
		defaultMomentModel: newMomentModel(conn),
	}
}

func (m *customMomentModel) withSession(session sqlx.Session) MomentModel {
	return NewMomentModel(sqlx.NewSqlConnFromSession(session))
}

func (m *customMomentModel) FindAll(ctx context.Context) ([]*Moment, error) {
	query := fmt.Sprintf("select %s from %s", momentRows, m.table)
	var resp []*Moment
	err := m.conn.QueryRowsCtx(ctx, &resp, query)
	return resp, err
}

func (m *customMomentModel) Count(ctx context.Context, condition string, values ...interface{}) (int, error) {
	query := fmt.Sprintf("select count(1) from %s where %s", m.table, condition)
	count := 0
	err := m.conn.QueryRowCtx(ctx, &count, query, values...)
	if err != nil {
		return 0, err
	}
	return count, nil
}
func (m *customMomentModel) FindByConditions(ctx context.Context, condition string, values ...interface{}) (*[]Moment, error) {
	query := fmt.Sprintf("select %s from %s where %s", momentRows, m.table, condition)
	var resp []Moment
	err := m.conn.QueryRowsCtx(ctx, &resp, query, values...)
	switch err {
	case nil:
		return &resp, nil
	default:
		return nil, err
	}
}

// MomentEx struct {
// 	MomentId     uint64    `db:"moment_id"`
// 	UserId       uint64    `db:"user_id"` // 用户id
// 	NickName     string    `db:"nick_name"`
// 	Avatar       string    `db:"avatar"`
// 	Content      string    `db:"content"` // 动态内容
// 	Images       string    `db:"images"`  // 动态图片
// 	Video        string    `db:"video"`
// 	Tags         string    `db:"tags"`
// 	Visibility   string    `db:"visibility"`
// 	IsTopic      int64     `db:"is_topic"`   // 话题动态 1是 0否
// 	CreatedAt    time.Time `db:"created_at"` // 创建时间
// 	UpdatedAt    time.Time `db:"updated_at"` // 修改时间
// 	FriendShow   int64     `db:"friendShow"` // 是否仅好友可见。0：否，1:是
// 	State        int64     `db:"state"`
// 	TopicId      int64     `db:"topic_id"` // 话题id
// 	Location     string    `db:"location"`
// 	Province     string    `db:"province"`
// 	City         string    `db:"city"`
// 	Region       string    `db:"region"`    // 区或者县
// 	Landmark     string    `db:"landmark"`  // 地标
// 	Longitude    string    `db:"longitude"` // 经度
// 	Latitude     string    `db:"latitude"`  // 纬度
// 	LikeCount    uint64    `db:"like_count"`
// 	CommentCount uint64    `db:"comment_count"`

// 	//===================
// 	Avatar string `db:"avatar"`
// 	NickName string `db:"nick_name"`

// }

// func (m *customMomentModel) FindExByConditions(ctx context.Context, condition string, values ...interface{}) (*[]Moment, error) {
// 	query := fmt.Sprintf("select a.*,b. from %s a join users b on a.user_id = b.user_id where %s",  m.table, condition)
// 	var resp []Moment
// 	err := m.conn.QueryRowsCtx(ctx, &resp, query, values...)
// 	switch err {
// 	case nil:
// 		return &resp, nil
// 	default:
// 		return nil, err
// 	}
// }

func (m *customMomentModel) FindByUserId(ctx context.Context, userId int64) (*Moment, error) {
	query := fmt.Sprintf("select %s from %s where `user_id` = ?", momentRows, m.table)
	var resp Moment
	err := m.conn.QueryRowCtx(ctx, &resp, query, userId)
	return &resp, err
}

func (m *customMomentModel) FindByCondition(ctx context.Context, condition string) (*Moment, error) {
	query := fmt.Sprintf("select %s from %s where %s", momentRows, m.table, condition)
	var resp Moment
	err := m.conn.QueryRowCtx(ctx, &resp, query)
	return &resp, err
}

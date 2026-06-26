package model

import "github.com/zeromicro/go-zero/core/stores/sqlx"

var _ SongPm5Model = (*customSongPm5Model)(nil)

type (
	// SongPm5Model is an interface to be customized, add more methods here,
	// and implement the added methods in customSongPm5Model.
	SongPm5Model interface {
		songPm5Model
		withSession(session sqlx.Session) SongPm5Model
	}

	customSongPm5Model struct {
		*defaultSongPm5Model
	}
)

// NewSongPm5Model returns a model for the database table.
func NewSongPm5Model(conn sqlx.SqlConn) SongPm5Model {
	return &customSongPm5Model{
		defaultSongPm5Model: newSongPm5Model(conn),
	}
}

func (m *customSongPm5Model) withSession(session sqlx.Session) SongPm5Model {
	return NewSongPm5Model(sqlx.NewSqlConnFromSession(session))
}

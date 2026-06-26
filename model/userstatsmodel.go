package model

import (
	"context"
	"fmt"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ UserStatsModel = (*customUserStatsModel)(nil)

type (
	// UserStatsModel is an interface to be customized, add more methods here,
	// and implement the added methods in customUserStatsModel.
	UserStatsModel interface {
		userStatsModel
		withSession(session sqlx.Session) UserStatsModel

		RecalculateVisitStatsByUser(ctx context.Context, userId int64, targetUserId int64) error
		RecalculateLikeStatsByUser(ctx context.Context, userId int64, targetUserId int64) error
		RecalculateMutualLikeStatsByUser(ctx context.Context, userId int64, targetUserId int64) error
		RecalculateCommentStatsByUser(ctx context.Context, userId int64, targetUserId int64) error
		RecalculateExamsStatsByUser(ctx context.Context, userId int64) error
		RecalculateCoursesStatsByUser(ctx context.Context, userId int64) error
		RecalculateMomentsStatsByUser(ctx context.Context, userId int64) error
	}

	customUserStatsModel struct {
		*defaultUserStatsModel
	}
)

// NewUserStatsModel returns a model for the database table.
func NewUserStatsModel(conn sqlx.SqlConn) UserStatsModel {
	return &customUserStatsModel{
		defaultUserStatsModel: newUserStatsModel(conn),
	}
}

func (m *customUserStatsModel) withSession(session sqlx.Session) UserStatsModel {
	return NewUserStatsModel(sqlx.NewSqlConnFromSession(session))
}

func (m *customUserStatsModel) RecalculateVisitStatsByUser(ctx context.Context, userId int64, targetUserId int64) error {
	// Count visits from userId to targetUserId
	var visitedCount uint64
	query := fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE `user_id` = ? AND `friend_id` = ?", "`user_visitors`")
	err := m.conn.QueryRowCtx(ctx, &visitedCount, query, userId, targetUserId)
	if err != nil {
		return fmt.Errorf("failed to count visits: %w", err)
	}

	// Count visits from targetUserId to userId
	var visitedmeCount uint64
	query = fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE `user_id` = ? AND `friend_id` = ?", "`user_visitors`")
	err = m.conn.QueryRowCtx(ctx, &visitedmeCount, query, targetUserId, userId)
	if err != nil {
		return fmt.Errorf("failed to count visits received: %w", err)
	}

	// Update user stats for userId
	userStats, err := m.FindOne(ctx, uint64(userId))
	if err != nil {
		return fmt.Errorf("failed to find user stats: %w", err)
	}
	userStats.VisitedCount = visitedCount
	userStats.VisitedmeCount = visitedmeCount
	err = m.Update(ctx, userStats)
	if err != nil {
		return fmt.Errorf("failed to update user stats: %w", err)
	}

	// Update user stats for targetUserId
	targetUserStats, err := m.FindOne(ctx, uint64(targetUserId))
	if err != nil {
		return fmt.Errorf("failed to find target user stats: %w", err)
	}
	targetUserStats.VisitedCount = visitedmeCount
	targetUserStats.VisitedmeCount = visitedCount
	err = m.Update(ctx, targetUserStats)
	if err != nil {
		return fmt.Errorf("failed to update target user stats: %w", err)
	}

	return nil
}

func (m *customUserStatsModel) RecalculateLikeStatsByUser(ctx context.Context, userId int64, targetUserId int64) error {
	// Count likes from userId to targetUserId (through user_friends table)
	var likedCount uint64
	query := fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE `user_id` = ? AND `friend_id` = ? AND `state` = 1", "`user_friends`")
	err := m.conn.QueryRowCtx(ctx, &likedCount, query, userId, targetUserId)
	if err != nil {
		return fmt.Errorf("failed to count likes: %w", err)
	}

	// Count likes from targetUserId to userId
	var likedmeCount uint64
	query = fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE `user_id` = ? AND `friend_id` = ? AND `state` = 1", "`user_friends`")
	err = m.conn.QueryRowCtx(ctx, &likedmeCount, query, targetUserId, userId)
	if err != nil {
		return fmt.Errorf("failed to count likes received: %w", err)
	}

	// Update user stats for userId
	userStats, err := m.FindOne(ctx, uint64(userId))
	if err != nil {
		return fmt.Errorf("failed to find user stats: %w", err)
	}
	userStats.LikedCount = likedCount
	userStats.LikedmeCount = likedmeCount
	err = m.Update(ctx, userStats)
	if err != nil {
		return fmt.Errorf("failed to update user stats: %w", err)
	}

	// Update user stats for targetUserId
	targetUserStats, err := m.FindOne(ctx, uint64(targetUserId))
	if err != nil {
		return fmt.Errorf("failed to find target user stats: %w", err)
	}
	targetUserStats.LikedCount = likedmeCount
	targetUserStats.LikedmeCount = likedCount
	err = m.Update(ctx, targetUserStats)
	if err != nil {
		return fmt.Errorf("failed to update target user stats: %w", err)
	}

	return nil
}

func (m *customUserStatsModel) RecalculateMutualLikeStatsByUser(ctx context.Context, userId int64, targetUserId int64) error {
	// Count mutual likes (both users like each other)
	var mutualLikeCount uint64
	query := fmt.Sprintf(`
		SELECT COUNT(*) FROM %s uf1 
		INNER JOIN %s uf2 ON uf1.user_id = uf2.friend_id AND uf1.friend_id = uf2.user_id 
		WHERE uf1.user_id = ? AND uf1.friend_id = ? AND uf1.state = 1 AND uf2.state = 1
	`, "`user_friends`", "`user_friends`")
	err := m.conn.QueryRowCtx(ctx, &mutualLikeCount, query, userId, targetUserId)
	if err != nil {
		return fmt.Errorf("failed to count mutual likes: %w", err)
	}

	// Update user stats for userId
	userStats, err := m.FindOne(ctx, uint64(userId))
	if err != nil {
		return fmt.Errorf("failed to find user stats: %w", err)
	}
	userStats.MutuallikedCount = mutualLikeCount
	err = m.Update(ctx, userStats)
	if err != nil {
		return fmt.Errorf("failed to update user stats: %w", err)
	}

	// Update user stats for targetUserId
	targetUserStats, err := m.FindOne(ctx, uint64(targetUserId))
	if err != nil {
		return fmt.Errorf("failed to find target user stats: %w", err)
	}
	targetUserStats.MutuallikedCount = mutualLikeCount
	err = m.Update(ctx, targetUserStats)
	if err != nil {
		return fmt.Errorf("failed to update target user stats: %w", err)
	}

	return nil
}

func (m *customUserStatsModel) RecalculateCommentStatsByUser(ctx context.Context, userId int64, targetUserId int64) error {
	// Count comments from userId to targetUserId's moments
	var commentCount uint64
	query := fmt.Sprintf(`
		SELECT COUNT(*) FROM %s mc 
		INNER JOIN %s m ON mc.moment_id = m.moment_id 
		WHERE mc.user_id = ? AND m.user_id = ?
	`, "`moment_comment`", "`moment`")
	err := m.conn.QueryRowCtx(ctx, &commentCount, query, userId, targetUserId)
	if err != nil {
		return fmt.Errorf("failed to count comments: %w", err)
	}

	// Count comments from targetUserId to userId's moments
	var commentmeCount uint64
	query = fmt.Sprintf(`
		SELECT COUNT(*) FROM %s mc 
		INNER JOIN %s m ON mc.moment_id = m.moment_id 
		WHERE mc.user_id = ? AND m.user_id = ?
	`, "`moment_comment`", "`moment`")
	err = m.conn.QueryRowCtx(ctx, &commentmeCount, query, targetUserId, userId)
	if err != nil {
		return fmt.Errorf("failed to count comments received: %w", err)
	}

	// Update user stats for userId
	userStats, err := m.FindOne(ctx, uint64(userId))
	if err != nil {
		return fmt.Errorf("failed to find user stats: %w", err)
	}
	userStats.CommentCount = commentCount
	userStats.CommentmeCount = commentmeCount
	err = m.Update(ctx, userStats)
	if err != nil {
		return fmt.Errorf("failed to update user stats: %w", err)
	}

	// Update user stats for targetUserId
	targetUserStats, err := m.FindOne(ctx, uint64(targetUserId))
	if err != nil {
		return fmt.Errorf("failed to find target user stats: %w", err)
	}
	targetUserStats.CommentCount = commentmeCount
	targetUserStats.CommentmeCount = commentCount
	err = m.Update(ctx, targetUserStats)
	if err != nil {
		return fmt.Errorf("failed to update target user stats: %w", err)
	}

	return nil
}

func (m *customUserStatsModel) RecalculateExamsStatsByUser(ctx context.Context, userId int64) error {
	// Count exams created by user
	var myTestsCount uint64
	query := fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE `user_id` = ?", "`exams`")
	err := m.conn.QueryRowCtx(ctx, &myTestsCount, query, userId)
	if err != nil {
		return fmt.Errorf("failed to count exams: %w", err)
	}

	// Count exam results for user (completed exams)
	var myCertsCount uint64
	query = fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE `user_id` = ? AND `is_temp` = 0", "`exams_result`")
	err = m.conn.QueryRowCtx(ctx, &myCertsCount, query, userId)
	if err != nil {
		return fmt.Errorf("failed to count exam results: %w", err)
	}

	// Update user stats
	userStats, err := m.FindOne(ctx, uint64(userId))
	if err != nil {
		return fmt.Errorf("failed to find user stats: %w", err)
	}
	userStats.MyTestsCount = myTestsCount
	userStats.MyCertsCount = myCertsCount
	err = m.Update(ctx, userStats)
	if err != nil {
		return fmt.Errorf("failed to update user stats: %w", err)
	}

	return nil
}

func (m *customUserStatsModel) RecalculateCoursesStatsByUser(ctx context.Context, userId int64) error {
	// Count knowledge courses acquired by user
	var coursesCount uint64
	query := fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE `user_id` = ?", "`user_acquire_course`")
	err := m.conn.QueryRowCtx(ctx, &coursesCount, query, userId)
	if err != nil {
		return fmt.Errorf("failed to count acquired courses: %w", err)
	}

	// Count knowledge points mastered by user
	var masteredCount uint64
	query = fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE `user_id` = ? AND `score` >= 80", "`knowledge_user_mastered`")
	err = m.conn.QueryRowCtx(ctx, &masteredCount, query, userId)
	if err != nil {
		return fmt.Errorf("failed to count mastered knowledge points: %w", err)
	}

	// Update user stats (using MyCertsCount for courses and MyTestsCount for mastered points)
	userStats, err := m.FindOne(ctx, uint64(userId))
	if err != nil {
		return fmt.Errorf("failed to find user stats: %w", err)
	}
	// Note: We're reusing existing fields since there's no specific course count field
	// MyCertsCount could represent completed courses, MyTestsCount could represent mastered points
	userStats.MyCertsCount = coursesCount
	userStats.MyTestsCount = masteredCount
	err = m.Update(ctx, userStats)
	if err != nil {
		return fmt.Errorf("failed to update user stats: %w", err)
	}

	return nil
}

func (m *customUserStatsModel) RecalculateMomentsStatsByUser(ctx context.Context, userId int64) error {
	// Count moments created by user
	var momentsCount uint64
	query := fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE `user_id` = ?", "`moment`")
	err := m.conn.QueryRowCtx(ctx, &momentsCount, query, userId)
	if err != nil {
		return fmt.Errorf("failed to count moments: %w", err)
	}

	// Count likes received on user's moments
	var momentslikeCount uint64
	query = fmt.Sprintf(`
		SELECT COUNT(*) FROM %s ml 
		INNER JOIN %s m ON ml.moment_id = m.moment_id 
		WHERE m.user_id = ?
	`, "`moment_like`", "`moment`")
	err = m.conn.QueryRowCtx(ctx, &momentslikeCount, query, userId)
	if err != nil {
		return fmt.Errorf("failed to count moment likes: %w", err)
	}

	// Count comments received on user's moments
	var momentsthumsupCount uint64
	query = fmt.Sprintf(`
		SELECT COUNT(*) FROM %s mc 
		INNER JOIN %s m ON mc.moment_id = m.moment_id 
		WHERE m.user_id = ?
	`, "`moment_comment`", "`moment`")
	err = m.conn.QueryRowCtx(ctx, &momentsthumsupCount, query, userId)
	if err != nil {
		return fmt.Errorf("failed to count moment comments: %w", err)
	}

	// Update user stats
	userStats, err := m.FindOne(ctx, uint64(userId))
	if err != nil {
		return fmt.Errorf("failed to find user stats: %w", err)
	}
	userStats.MomentsCount = momentsCount
	userStats.MomentslikeCount = momentslikeCount
	userStats.MomentsthumsupCount = momentsthumsupCount
	err = m.Update(ctx, userStats)
	if err != nil {
		return fmt.Errorf("failed to update user stats: %w", err)
	}

	return nil
}

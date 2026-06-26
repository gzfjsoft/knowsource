package model

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ BlogModel = (*customBlogModel)(nil)

type (
	// BlogModel is an interface to be customized, add more methods here,
	// and implement the added methods in customBlogModel.
	BlogModel interface {
		blogModel
		withSession(session sqlx.Session) BlogModel
		FindByUserId(ctx context.Context, userId uint64, title string, tags []string, categories []string, isPublished int64, page, pageSize int64) ([]*Blog, int64, error)
		FindAll(ctx context.Context, title string, tags []string, categories []string, isPublished int64, page, pageSize int64) ([]*Blog, int64, error)
		FindAllForAdmin(ctx context.Context, title string, tags []string, categories []string, isPublished int64, page, pageSize int64) ([]*Blog, int64, error)
		FindPublicList(ctx context.Context, keyword string, page, pageSize int64) ([]*BlogRecord, int64, error)
		FindPublicBySlug(ctx context.Context, slug string) (*BlogRecord, error)
		FindAdminList(ctx context.Context, keyword string, isPublished, page, pageSize int64) ([]*BlogRecord, int64, error)
		FindAdminById(ctx context.Context, id int64) (*BlogRecord, error)
		ExistsAlias(ctx context.Context, alias string, excludeId int64) (bool, error)
		CreateAdmin(ctx context.Context, in *BlogAdminMutation) (int64, error)
		UpdateAdmin(ctx context.Context, in *BlogAdminMutation) error
		DeleteAdmin(ctx context.Context, id int64) error
	}

	customBlogModel struct {
		*defaultBlogModel
	}

	BlogRecord struct {
		Id          int64          `db:"id"`
		Title       string         `db:"title"`
		Alias       sql.NullString `db:"alias"`
		Summary     string         `db:"summary"`
		Content     string         `db:"content"`
		Tags        string         `db:"tags"`
		Categories  string         `db:"categories"`
		Authors     string         `db:"authors"`
		Banner      string         `db:"banner"`
		IsPublished int64          `db:"is_published"`
		CreatedAt   time.Time      `db:"created_at"`
		UpdatedAt   time.Time      `db:"updated_at"`
	}

	BlogAdminMutation struct {
		Id          int64
		Title       string
		Alias       string
		Summary     string
		Content     string
		Tags        string
		Categories  string
		Authors     string
		Banner      string
		IsPublished int64
	}
)

// NewBlogModel returns a model for the database table.
func NewBlogModel(conn sqlx.SqlConn) BlogModel {
	return &customBlogModel{
		defaultBlogModel: newBlogModel(conn),
	}
}

func (m *customBlogModel) withSession(session sqlx.Session) BlogModel {
	return NewBlogModel(sqlx.NewSqlConnFromSession(session))
}

func (m *customBlogModel) FindByUserId(ctx context.Context, userId uint64, title string, tags []string, categories []string, isPublished int64, page, pageSize int64) ([]*Blog, int64, error) {
	// Build query conditions
	var conditions []string
	var args []interface{}

	// Add user ID condition
	conditions = append(conditions, "user_id = ?")
	args = append(args, userId)

	if title != "" {
		conditions = append(conditions, "title LIKE ?")
		args = append(args, fmt.Sprintf("%%%s%%", title))
	}

	if len(tags) > 0 {
		for _, tag := range tags {
			conditions = append(conditions, "tags LIKE ?")
			args = append(args, fmt.Sprintf("%%%s%%", tag))
		}
	}

	if len(categories) > 0 {
		for _, category := range categories {
			conditions = append(conditions, "categories LIKE ?")
			args = append(args, fmt.Sprintf("%%%s%%", category))
		}
	}

	if isPublished != -1 {
		conditions = append(conditions, "is_published = ?")
		args = append(args, isPublished)
	}

	// Build WHERE clause
	whereClause := ""
	if len(conditions) > 0 {
		whereClause = "WHERE " + strings.Join(conditions, " AND ")
	}

	// Build query
	query := fmt.Sprintf("SELECT %s FROM %s %s ORDER BY created_at DESC LIMIT ? OFFSET ?", blogRows, m.table, whereClause)
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM %s %s", m.table, whereClause)

	// Add pagination parameters
	args = append(args, pageSize, (page-1)*pageSize)

	// Execute count query
	var total int64
	err := m.conn.QueryRowCtx(ctx, &total, countQuery, args[:len(args)-2]...)
	if err != nil {
		return nil, 0, err
	}

	// Execute main query
	var blogs []*Blog
	err = m.conn.QueryRowsCtx(ctx, &blogs, query, args...)
	if err != nil {
		return nil, 0, err
	}

	return blogs, total, nil
}

func (m *customBlogModel) FindAll(ctx context.Context, title string, tags []string, categories []string, isPublished int64, page, pageSize int64) ([]*Blog, int64, error) {
	// Build query conditions
	var conditions []string
	var args []interface{}

	if title != "" {
		conditions = append(conditions, "title LIKE ?")
		args = append(args, fmt.Sprintf("%%%s%%", title))
	}

	if len(tags) > 0 {
		for _, tag := range tags {
			conditions = append(conditions, "tags LIKE ?")
			args = append(args, fmt.Sprintf("%%%s%%", tag))
		}
	}

	if len(categories) > 0 {
		for _, category := range categories {
			conditions = append(conditions, "categories LIKE ?")
			args = append(args, fmt.Sprintf("%%%s%%", category))
		}
	}

	if isPublished != -1 {
		conditions = append(conditions, "is_published = ?")
		args = append(args, isPublished)
	}

	// Build WHERE clause
	whereClause := ""
	if len(conditions) > 0 {
		whereClause = "WHERE " + strings.Join(conditions, " AND ")
	}

	// Build query
	query := fmt.Sprintf("SELECT %s FROM %s %s ORDER BY created_at DESC LIMIT ? OFFSET ?", blogRows, m.table, whereClause)
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM %s %s", m.table, whereClause)

	// Add pagination parameters
	args = append(args, pageSize, (page-1)*pageSize)

	// Execute count query
	var total int64
	err := m.conn.QueryRowCtx(ctx, &total, countQuery, args[:len(args)-2]...)
	if err != nil {
		return nil, 0, err
	}

	// Execute main query
	var blogs []*Blog
	err = m.conn.QueryRowsCtx(ctx, &blogs, query, args...)
	if err != nil {
		return nil, 0, err
	}

	return blogs, total, nil
}

func (m *customBlogModel) FindAllForAdmin(ctx context.Context, title string, tags []string, categories []string, isPublished int64, page, pageSize int64) ([]*Blog, int64, error) {
	// Build query conditions
	var conditions []string
	var args []interface{}

	if title != "" {
		conditions = append(conditions, "title LIKE ?")
		args = append(args, fmt.Sprintf("%%%s%%", title))
	}

	if len(tags) > 0 {
		for _, tag := range tags {
			conditions = append(conditions, "tags LIKE ?")
			args = append(args, fmt.Sprintf("%%%s%%", tag))
		}
	}

	if len(categories) > 0 {
		for _, category := range categories {
			conditions = append(conditions, "categories LIKE ?")
			args = append(args, fmt.Sprintf("%%%s%%", category))
		}
	}

	if isPublished != 0 {
		conditions = append(conditions, "is_published = ?")
		args = append(args, isPublished)
	}

	// Build WHERE clause
	whereClause := ""
	if len(conditions) > 0 {
		whereClause = "WHERE " + strings.Join(conditions, " AND ")
	}

	// Build query
	query := fmt.Sprintf("SELECT %s FROM %s %s ORDER BY created_at DESC LIMIT ? OFFSET ?", blogRows, m.table, whereClause)
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM %s %s", m.table, whereClause)

	// Add pagination parameters
	args = append(args, pageSize, (page-1)*pageSize)

	// Execute count query
	var total int64
	err := m.conn.QueryRowCtx(ctx, &total, countQuery, args[:len(args)-2]...)
	if err != nil {
		return nil, 0, err
	}

	// Execute main query
	var blogs []*Blog
	err = m.conn.QueryRowsCtx(ctx, &blogs, query, args...)
	if err != nil {
		return nil, 0, err
	}

	return blogs, total, nil
}

func normalizePage(page, pageSize int64) (int64, int64) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	return page, pageSize
}

func normalizeAlias(alias string) string {
	return strings.TrimSpace(alias)
}

func normalizeLike(keyword string) string {
	keyword = strings.TrimSpace(keyword)
	if keyword == "" {
		return ""
	}
	return "%" + keyword + "%"
}

func (m *customBlogModel) FindPublicList(ctx context.Context, keyword string, page, pageSize int64) ([]*BlogRecord, int64, error) {
	page, pageSize = normalizePage(page, pageSize)
	where := "WHERE is_published = 1"
	args := make([]interface{}, 0)
	if like := normalizeLike(keyword); like != "" {
		where += " AND (title LIKE ? OR summary LIKE ? OR alias LIKE ?)"
		args = append(args, like, like, like)
	}
	countQuery := fmt.Sprintf("SELECT COUNT(1) FROM %s %s", m.table, where)
	var total int64
	if err := m.conn.QueryRowCtx(ctx, &total, countQuery, args...); err != nil {
		return nil, 0, err
	}
	query := fmt.Sprintf("SELECT id,title,alias,summary,content,tags,categories,authors,banner,is_published,created_at,updated_at FROM %s %s ORDER BY created_at DESC LIMIT ? OFFSET ?", m.table, where)
	args = append(args, pageSize, (page-1)*pageSize)
	out := make([]*BlogRecord, 0)
	if err := m.conn.QueryRowsCtx(ctx, &out, query, args...); err != nil {
		return nil, 0, err
	}
	return out, total, nil
}

func (m *customBlogModel) FindPublicBySlug(ctx context.Context, slug string) (*BlogRecord, error) {
	slug = strings.TrimSpace(slug)
	if slug == "" {
		return nil, ErrNotFound
	}
	var (
		query string
		args  []interface{}
	)
	if id, err := strconv.ParseInt(slug, 10, 64); err == nil && id > 0 {
		query = fmt.Sprintf("SELECT id,title,alias,summary,content,tags,categories,authors,banner,is_published,created_at,updated_at FROM %s WHERE id = ? AND is_published = 1 LIMIT 1", m.table)
		args = []interface{}{id}
	} else {
		query = fmt.Sprintf("SELECT id,title,alias,summary,content,tags,categories,authors,banner,is_published,created_at,updated_at FROM %s WHERE alias = ? AND is_published = 1 LIMIT 1", m.table)
		args = []interface{}{slug}
	}
	var out BlogRecord
	err := m.conn.QueryRowCtx(ctx, &out, query, args...)
	switch err {
	case nil:
		return &out, nil
	case sqlx.ErrNotFound:
		return nil, ErrNotFound
	default:
		return nil, err
	}
}

func (m *customBlogModel) FindAdminList(ctx context.Context, keyword string, isPublished, page, pageSize int64) ([]*BlogRecord, int64, error) {
	page, pageSize = normalizePage(page, pageSize)
	where := "WHERE 1=1"
	args := make([]interface{}, 0)
	if like := normalizeLike(keyword); like != "" {
		where += " AND (title LIKE ? OR summary LIKE ? OR alias LIKE ?)"
		args = append(args, like, like, like)
	}
	if isPublished == 0 || isPublished == 1 {
		where += " AND is_published = ?"
		args = append(args, isPublished)
	}
	countQuery := fmt.Sprintf("SELECT COUNT(1) FROM %s %s", m.table, where)
	var total int64
	if err := m.conn.QueryRowCtx(ctx, &total, countQuery, args...); err != nil {
		return nil, 0, err
	}
	query := fmt.Sprintf("SELECT id,title,alias,summary,content,tags,categories,authors,banner,is_published,created_at,updated_at FROM %s %s ORDER BY created_at DESC LIMIT ? OFFSET ?", m.table, where)
	args = append(args, pageSize, (page-1)*pageSize)
	out := make([]*BlogRecord, 0)
	if err := m.conn.QueryRowsCtx(ctx, &out, query, args...); err != nil {
		return nil, 0, err
	}
	return out, total, nil
}

func (m *customBlogModel) FindAdminById(ctx context.Context, id int64) (*BlogRecord, error) {
	query := fmt.Sprintf("SELECT id,title,alias,summary,content,tags,categories,authors,banner,is_published,created_at,updated_at FROM %s WHERE id = ? LIMIT 1", m.table)
	var out BlogRecord
	err := m.conn.QueryRowCtx(ctx, &out, query, id)
	switch err {
	case nil:
		return &out, nil
	case sqlx.ErrNotFound:
		return nil, ErrNotFound
	default:
		return nil, err
	}
}

func (m *customBlogModel) ExistsAlias(ctx context.Context, alias string, excludeId int64) (bool, error) {
	alias = normalizeAlias(alias)
	if alias == "" {
		return false, nil
	}
	query := fmt.Sprintf("SELECT COUNT(1) FROM %s WHERE alias = ?", m.table)
	args := []interface{}{alias}
	if excludeId > 0 {
		query += " AND id <> ?"
		args = append(args, excludeId)
	}
	var total int64
	if err := m.conn.QueryRowCtx(ctx, &total, query, args...); err != nil {
		return false, err
	}
	return total > 0, nil
}

func (m *customBlogModel) CreateAdmin(ctx context.Context, in *BlogAdminMutation) (int64, error) {
	if in == nil {
		return 0, fmt.Errorf("blog mutation is nil")
	}
	var aliasVal interface{}
	if a := normalizeAlias(in.Alias); a != "" {
		aliasVal = a
	}
	query := fmt.Sprintf("INSERT INTO %s (title,sort_order,tags,categories,banner,twitter_author,user_id,authors,summary,content,is_approved,is_published,uuid,alias,created_at,updated_at) VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,NOW(),NOW())", m.table)
	ret, err := m.conn.ExecCtx(ctx, query,
		strings.TrimSpace(in.Title), 0, strings.TrimSpace(in.Tags), strings.TrimSpace(in.Categories),
		strings.TrimSpace(in.Banner), "", 0, strings.TrimSpace(in.Authors),
		strings.TrimSpace(in.Summary), in.Content, 1, in.IsPublished, "", aliasVal,
	)
	if err != nil {
		return 0, err
	}
	id, err := ret.LastInsertId()
	if err != nil {
		return 0, err
	}
	return id, nil
}

func (m *customBlogModel) UpdateAdmin(ctx context.Context, in *BlogAdminMutation) error {
	if in == nil || in.Id <= 0 {
		return fmt.Errorf("invalid blog mutation")
	}
	var aliasVal interface{}
	if a := normalizeAlias(in.Alias); a != "" {
		aliasVal = a
	}
	query := fmt.Sprintf("UPDATE %s SET title=?, tags=?, categories=?, banner=?, authors=?, summary=?, content=?, is_published=?, alias=?, updated_at=NOW() WHERE id=?", m.table)
	_, err := m.conn.ExecCtx(ctx, query,
		strings.TrimSpace(in.Title), strings.TrimSpace(in.Tags), strings.TrimSpace(in.Categories),
		strings.TrimSpace(in.Banner), strings.TrimSpace(in.Authors), strings.TrimSpace(in.Summary),
		in.Content, in.IsPublished, aliasVal, in.Id,
	)
	return err
}

func (m *customBlogModel) DeleteAdmin(ctx context.Context, id int64) error {
	query := fmt.Sprintf("DELETE FROM %s WHERE id = ?", m.table)
	_, err := m.conn.ExecCtx(ctx, query, id)
	return err
}

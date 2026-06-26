package model

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ RawDocumentsModel = (*customRawDocumentsModel)(nil)

type (
	RawDocumentsSearchResult struct {
		Id           int64     `db:"id"`
		DocumentCode string    `db:"document_code"`
		FileName     string    `db:"file_name"`
		Tag          string    `db:"tag"`
		Content      string    `db:"content"`
		Snippet      string    `db:"snippet"`
		CreatedAt    time.Time `db:"created_at"`
		UpdatedAt    time.Time `db:"updated_at"`
		IsAudit      int64     `db:"is_audit"`
		IsToMd       int64     `db:"is_to_md"`
		IsToAi       int64     `db:"is_to_ai"`
		Status       string    `db:"status"`
		StatusMsg    string    `db:"status_msg"`
	}

	// RawDocumentsModel is an interface to be customized, add more methods here,
	// and implement the added methods in customRawDocumentsModel.
	RawDocumentsModel interface {
		rawDocumentsModel
		WithSession(session sqlx.Session) RawDocumentsModel
		FindOneByClientId(ctx context.Context, clientId string, id int64) (*RawDocuments, error)
		DeleteByClientId(ctx context.Context, clientId string, id int64) error
		FindByDocumentCode(ctx context.Context, clientId string, documentCode string, fileName string, tag string, isAudit string, offset, limit int64) ([]*RawDocuments, error)
		CountByDocumentCode(ctx context.Context, clientId string, documentCode string, fileName string, tag string, isAudit string) (int64, error)
		SearchByKeyword(ctx context.Context, clientId string, keyword string, documentCode string, tag string, isAudit string, offset, limit int64) ([]*RawDocumentsSearchResult, error)
		SearchByKeywordAll(ctx context.Context, clientId string, keyword string, documentCode string, tag string, isAudit string) ([]*RawDocumentsSearchResult, error)
		CountByKeyword(ctx context.Context, clientId string, keyword string, documentCode string, tag string, isAudit string) (int64, error)
		FindByMD5AndDocumentCode(ctx context.Context, clientId string, fileMd5 string, documentCode string) (*RawDocuments, error)
		FindByFileNameAndDocumentCode(ctx context.Context, clientId string, fileName string, documentCode string) (*RawDocuments, error)
		FindByFileName(ctx context.Context, clientId string, fileName string) (*RawDocuments, error)
		FindDistinctTags(ctx context.Context, clientId string, documentCode string) ([]string, error)
		UpdateIsToAi(ctx context.Context, clientId string, fileName string, documentCode string, value int64) error
		FindAll(ctx context.Context, clientId string) ([]*RawDocuments, error)
		// CountAuditedByTag 统计使用该标签的 raw_documents 数量（无论是否审核），用于删除标签前校验
		CountAuditedByTag(ctx context.Context, clientId string, tag string) (int64, error)
		// ClearTag 将该标签从所有 raw_documents 中清空（删除标签时一并清理未审核文档上的该标签）
		ClearTag(ctx context.Context, clientId string, tag string) error
		// UpdateStatusAndMsg 仅更新 status、status_msg（异步任务状态流转专用）
		UpdateStatusAndMsg(ctx context.Context, clientId string, id int64, status, statusMsg string) (int64, error)
		// UpdateStatusOnly 仅更新 status（兼容尚未执行 status_msg 迁移的库）
		UpdateStatusOnly(ctx context.Context, clientId string, id int64, status string) (int64, error)
		// ClearStatusMsg 清空状态说明（开始新异步任务时清除上次失败原因）
		ClearStatusMsg(ctx context.Context, clientId string, id int64) (int64, error)
		// ClearAuditFields 清除审核标记（中断入库/取消审核）
		ClearAuditFields(ctx context.Context, clientId string, id int64) error
	}

	customRawDocumentsModel struct {
		*defaultRawDocumentsModel
	}
)

// NewRawDocumentsModel returns a model for the database table.
func NewRawDocumentsModel(conn sqlx.SqlConn) RawDocumentsModel {
	return &customRawDocumentsModel{
		defaultRawDocumentsModel: newRawDocumentsModel(conn),
	}
}

func (m *customRawDocumentsModel) WithSession(session sqlx.Session) RawDocumentsModel {
	return NewRawDocumentsModel(sqlx.NewSqlConnFromSession(session))
}

func appendRawDocClientFilter(clientId string, where *[]string, args *[]interface{}) {
	if strings.TrimSpace(clientId) == "" {
		return
	}
	*where = append(*where, "client_id = ?")
	*args = append(*args, clientId)
}

func (m *customRawDocumentsModel) FindOneByClientId(ctx context.Context, clientId string, id int64) (*RawDocuments, error) {
	query := fmt.Sprintf("SELECT %s FROM %s WHERE `id` = ? AND `client_id` = ? LIMIT 1", rawDocumentsRows, m.table)
	var resp RawDocuments
	err := m.conn.QueryRowCtx(ctx, &resp, query, id, clientId)
	switch err {
	case nil:
		return &resp, nil
	case sqlx.ErrNotFound:
		return nil, ErrNotFound
	default:
		return nil, err
	}
}

func (m *customRawDocumentsModel) DeleteByClientId(ctx context.Context, clientId string, id int64) error {
	if strings.TrimSpace(clientId) == "" {
		return m.Delete(ctx, id)
	}
	query := fmt.Sprintf("DELETE FROM %s WHERE `id` = ? AND `client_id` = ?", m.table)
	_, err := m.conn.ExecCtx(ctx, query, id, clientId)
	return err
}

func (m *customRawDocumentsModel) SearchByKeyword(ctx context.Context, clientId string, keyword string, documentCode string, tag string, isAudit string, offset, limit int64) ([]*RawDocumentsSearchResult, error) {
	keyword = strings.TrimSpace(keyword)
	if keyword == "" {
		return []*RawDocumentsSearchResult{}, nil
	}

	// 先尝试 MySQL FULLTEXT（需要 FULLTEXT 索引）
	query := fmt.Sprintf(`
SELECT
  id,
  document_code,
  file_name,
  tag,
  CASE
    WHEN LOCATE(?, content) > 0 THEN SUBSTRING(content, GREATEST(LOCATE(?, content) - 400, 1), 2000)
    WHEN LOCATE(?, file_name) > 0 THEN file_name
    ELSE SUBSTRING(content, 1, 2000)
  END AS snippet,
  created_at,
  updated_at,
  is_audit,
  is_to_md,
  is_to_ai,
  status,
  status_msg
FROM %s
`, m.table)
	prefixArgs := []interface{}{keyword, keyword, keyword}
	whereClauses := []string{}
	condArgs := []interface{}{}
	appendRawDocClientFilter(clientId, &whereClauses, &condArgs)
	whereClauses = append(whereClauses, "MATCH(content, file_name) AGAINST (? IN NATURAL LANGUAGE MODE)")
	condArgs = append(condArgs, keyword)

	if documentCode != "" {
		whereClauses = append(whereClauses, "document_code = ?")
		condArgs = append(condArgs, documentCode)
	}
	if tag != "" {
		whereClauses = append(whereClauses, "tag = ?")
		condArgs = append(condArgs, tag)
	}
	if isAudit != "" {
		whereClauses = append(whereClauses, "is_audit = ?")
		condArgs = append(condArgs, isAudit)
	}

	if len(whereClauses) > 0 {
		query += " WHERE " + whereClauses[0]
		for i := 1; i < len(whereClauses); i++ {
			query += " AND " + whereClauses[i]
		}
	}
	query += " ORDER BY created_at DESC LIMIT ? OFFSET ?"
	args := append(prefixArgs, condArgs...)
	args = append(args, limit, offset)

	var rows []*RawDocumentsSearchResult
	err := m.conn.QueryRowsCtx(ctx, &rows, query, args...)
	if err == nil {
		return rows, nil
	}

	// 如果没有 FULLTEXT 索引，则降级到 LIKE
	errMsgLower := strings.ToLower(err.Error())
	if !strings.Contains(errMsgLower, "fulltext") {
		return nil, err
	}

	like := "%" + keyword + "%"
	query = fmt.Sprintf(`
SELECT
  id,
  document_code,
  file_name,
  tag,
  CASE
    WHEN LOCATE(?, content) > 0 THEN SUBSTRING(content, GREATEST(LOCATE(?, content) - 400, 1), 2000)
    WHEN LOCATE(?, file_name) > 0 THEN file_name
    ELSE SUBSTRING(content, 1, 2000)
  END AS snippet,
  created_at,
  updated_at,
  is_audit,
  is_to_md,
  is_to_ai,
  status,
  status_msg
FROM %s
`, m.table)
	prefixArgs = []interface{}{keyword, keyword, keyword}
	whereClauses = []string{}
	condArgs = []interface{}{}
	appendRawDocClientFilter(clientId, &whereClauses, &condArgs)
	whereClauses = append(whereClauses, "(content LIKE ? OR file_name LIKE ?)")
	condArgs = append(condArgs, like, like)

	if documentCode != "" {
		whereClauses = append(whereClauses, "document_code = ?")
		condArgs = append(condArgs, documentCode)
	}
	if tag != "" {
		whereClauses = append(whereClauses, "tag = ?")
		condArgs = append(condArgs, tag)
	}
	if isAudit != "" {
		whereClauses = append(whereClauses, "is_audit = ?")
		condArgs = append(condArgs, isAudit)
	}

	query += " WHERE " + whereClauses[0]
	for i := 1; i < len(whereClauses); i++ {
		query += " AND " + whereClauses[i]
	}
	query += " ORDER BY created_at DESC LIMIT ? OFFSET ?"
	args = append(prefixArgs, condArgs...)
	args = append(args, limit, offset)

	rows = []*RawDocumentsSearchResult{}
	err = m.conn.QueryRowsCtx(ctx, &rows, query, args...)
	return rows, err
}

func (m *customRawDocumentsModel) SearchByKeywordAll(ctx context.Context, clientId string, keyword string, documentCode string, tag string, isAudit string) ([]*RawDocumentsSearchResult, error) {
	keyword = strings.TrimSpace(keyword)
	if keyword == "" {
		return []*RawDocumentsSearchResult{}, nil
	}

	query := fmt.Sprintf(`
SELECT
  id,
  document_code,
  file_name,
  tag,
  content,
  CASE
    WHEN LOCATE(?, content) > 0 THEN SUBSTRING(content, GREATEST(LOCATE(?, content) - 400, 1), 2000)
    WHEN LOCATE(?, file_name) > 0 THEN file_name
    ELSE SUBSTRING(content, 1, 2000)
  END AS snippet,
  created_at,
  updated_at,
  is_audit,
  is_to_md,
  is_to_ai,
  status,
  status_msg
FROM %s
`, m.table)
	prefixArgs := []interface{}{keyword, keyword, keyword}
	whereClauses := []string{}
	condArgs := []interface{}{}
	appendRawDocClientFilter(clientId, &whereClauses, &condArgs)
	whereClauses = append(whereClauses, "MATCH(content, file_name) AGAINST (? IN NATURAL LANGUAGE MODE)")
	condArgs = append(condArgs, keyword)

	if documentCode != "" {
		whereClauses = append(whereClauses, "document_code = ?")
		condArgs = append(condArgs, documentCode)
	}
	if tag != "" {
		whereClauses = append(whereClauses, "tag = ?")
		condArgs = append(condArgs, tag)
	}
	if isAudit != "" {
		whereClauses = append(whereClauses, "is_audit = ?")
		condArgs = append(condArgs, isAudit)
	}

	if len(whereClauses) > 0 {
		query += " WHERE " + whereClauses[0]
		for i := 1; i < len(whereClauses); i++ {
			query += " AND " + whereClauses[i]
		}
	}
	query += " ORDER BY created_at DESC"
	args := append(prefixArgs, condArgs...)

	var rows []*RawDocumentsSearchResult
	err := m.conn.QueryRowsCtx(ctx, &rows, query, args...)
	if err == nil {
		return rows, nil
	}

	// 如果没有 FULLTEXT 索引，则降级到 LIKE
	errMsgLower := strings.ToLower(err.Error())
	if !strings.Contains(errMsgLower, "fulltext") {
		return nil, err
	}

	like := "%" + keyword + "%"
	query = fmt.Sprintf(`
SELECT
  id,
  document_code,
  file_name,
  tag,
  content,
  CASE
    WHEN LOCATE(?, content) > 0 THEN SUBSTRING(content, GREATEST(LOCATE(?, content) - 400, 1), 2000)
    WHEN LOCATE(?, file_name) > 0 THEN file_name
    ELSE SUBSTRING(content, 1, 2000)
  END AS snippet,
  created_at,
  updated_at,
  is_audit,
  is_to_md,
  is_to_ai,
  status,
  status_msg
FROM %s
`, m.table)
	whereClauses = []string{}
	condArgs = []interface{}{}
	appendRawDocClientFilter(clientId, &whereClauses, &condArgs)
	whereClauses = append(whereClauses, "(content LIKE ? OR file_name LIKE ?)")
	condArgs = append(condArgs, like, like)

	if documentCode != "" {
		whereClauses = append(whereClauses, "document_code = ?")
		condArgs = append(condArgs, documentCode)
	}
	if tag != "" {
		whereClauses = append(whereClauses, "tag = ?")
		condArgs = append(condArgs, tag)
	}
	if isAudit != "" {
		whereClauses = append(whereClauses, "is_audit = ?")
		condArgs = append(condArgs, isAudit)
	}

	query += " WHERE " + whereClauses[0]
	for i := 1; i < len(whereClauses); i++ {
		query += " AND " + whereClauses[i]
	}
	query += " ORDER BY created_at DESC"
	args = append(prefixArgs, condArgs...)

	rows = []*RawDocumentsSearchResult{}
	err = m.conn.QueryRowsCtx(ctx, &rows, query, args...)
	return rows, err
}

func (m *customRawDocumentsModel) CountByKeyword(ctx context.Context, clientId string, keyword string, documentCode string, tag string, isAudit string) (int64, error) {
	keyword = strings.TrimSpace(keyword)
	if keyword == "" {
		return 0, nil
	}

	query := fmt.Sprintf("SELECT COUNT(*) FROM %s", m.table)
	whereClauses := []string{}
	args := []interface{}{}
	appendRawDocClientFilter(clientId, &whereClauses, &args)
	whereClauses = append(whereClauses, "MATCH(content, file_name) AGAINST (? IN NATURAL LANGUAGE MODE)")
	args = append(args, keyword)

	if documentCode != "" {
		whereClauses = append(whereClauses, "document_code = ?")
		args = append(args, documentCode)
	}
	if tag != "" {
		whereClauses = append(whereClauses, "tag = ?")
		args = append(args, tag)
	}
	if isAudit != "" {
		whereClauses = append(whereClauses, "is_audit = ?")
		args = append(args, isAudit)
	}

	if len(whereClauses) > 0 {
		query += " WHERE " + whereClauses[0]
		for i := 1; i < len(whereClauses); i++ {
			query += " AND " + whereClauses[i]
		}
	}

	var count int64
	err := m.conn.QueryRowCtx(ctx, &count, query, args...)
	if err == nil {
		return count, nil
	}

	// 如果没有 FULLTEXT 索引，则降级到 LIKE
	errMsgLower := strings.ToLower(err.Error())
	if !strings.Contains(errMsgLower, "fulltext") {
		return 0, err
	}

	like := "%" + keyword + "%"
	query = fmt.Sprintf("SELECT COUNT(*) FROM %s", m.table)
	whereClauses = []string{}
	args = []interface{}{}
	appendRawDocClientFilter(clientId, &whereClauses, &args)
	whereClauses = append(whereClauses, "(content LIKE ? OR file_name LIKE ?)")
	args = append(args, like, like)

	if documentCode != "" {
		whereClauses = append(whereClauses, "document_code = ?")
		args = append(args, documentCode)
	}
	if tag != "" {
		whereClauses = append(whereClauses, "tag = ?")
		args = append(args, tag)
	}
	if isAudit != "" {
		whereClauses = append(whereClauses, "is_audit = ?")
		args = append(args, isAudit)
	}

	query += " WHERE " + whereClauses[0]
	for i := 1; i < len(whereClauses); i++ {
		query += " AND " + whereClauses[i]
	}

	count = 0
	err = m.conn.QueryRowCtx(ctx, &count, query, args...)
	return count, err
}

func (m *customRawDocumentsModel) UpdateIsToAi(ctx context.Context, clientId string, fileName string, documentCode string, value int64) error {
	if strings.TrimSpace(clientId) == "" {
		query := fmt.Sprintf("UPDATE %s SET is_to_ai = ? WHERE file_name = ? AND document_code = ?", m.table)
		_, err := m.conn.ExecCtx(ctx, query, value, fileName, documentCode)
		return err
	}
	query := fmt.Sprintf("UPDATE %s SET is_to_ai = ? WHERE client_id = ? AND file_name = ? AND document_code = ?", m.table)
	_, err := m.conn.ExecCtx(ctx, query, value, clientId, fileName, documentCode)
	return err
}

func (m *customRawDocumentsModel) FindByDocumentCode(ctx context.Context, clientId string, documentCode string, fileName string, tag string, isAudit string, offset, limit int64) ([]*RawDocuments, error) {
	query := fmt.Sprintf("SELECT %s FROM %s", rawDocumentsRows, m.table)
	args := []interface{}{}
	whereClauses := []string{}
	appendRawDocClientFilter(clientId, &whereClauses, &args)
	if documentCode != "" {
		whereClauses = append(whereClauses, "document_code = ?")
		args = append(args, documentCode)
	}

	if fileName != "" {
		whereClauses = append(whereClauses, "file_name LIKE ?")
		args = append(args, "%"+fileName+"%")
	}

	if tag != "" {
		whereClauses = append(whereClauses, "tag = ?")
		args = append(args, tag)
	}

	if isAudit != "" {
		whereClauses = append(whereClauses, "is_audit = ?")
		args = append(args, isAudit)
	}

	if len(whereClauses) > 0 {
		query += " WHERE " + whereClauses[0]
		for i := 1; i < len(whereClauses); i++ {
			query += " AND " + whereClauses[i]
		}
	}

	query += " ORDER BY created_at DESC LIMIT ? OFFSET ?"
	args = append(args, limit, offset)

	var documents []*RawDocuments
	err := m.conn.QueryRowsCtx(ctx, &documents, query, args...)
	return documents, err
}

func (m *customRawDocumentsModel) CountByDocumentCode(ctx context.Context, clientId string, documentCode string, fileName string, tag string, isAudit string) (int64, error) {
	query := fmt.Sprintf("SELECT COUNT(*) FROM %s", m.table)
	args := []interface{}{}
	whereClauses := []string{}
	appendRawDocClientFilter(clientId, &whereClauses, &args)

	if documentCode != "" {
		whereClauses = append(whereClauses, "document_code = ?")
		args = append(args, documentCode)
	}

	if fileName != "" {
		whereClauses = append(whereClauses, "file_name LIKE ?")
		args = append(args, "%"+fileName+"%")
	}

	if tag != "" {
		whereClauses = append(whereClauses, "tag = ?")
		args = append(args, tag)
	}

	if isAudit != "" {
		whereClauses = append(whereClauses, "is_audit = ?")
		args = append(args, isAudit)
	}

	if len(whereClauses) > 0 {
		query += " WHERE " + whereClauses[0]
		for i := 1; i < len(whereClauses); i++ {
			query += " AND " + whereClauses[i]
		}
	}

	var count int64
	err := m.conn.QueryRowCtx(ctx, &count, query, args...)
	return count, err
}

func (m *customRawDocumentsModel) FindByMD5AndDocumentCode(ctx context.Context, clientId string, fileMd5 string, documentCode string) (*RawDocuments, error) {
	var query string
	var args []interface{}
	if strings.TrimSpace(clientId) == "" {
		query = fmt.Sprintf("SELECT %s FROM %s WHERE file_md5 = ? AND document_code = ? LIMIT 1", rawDocumentsRows, m.table)
		args = []interface{}{fileMd5, documentCode}
	} else {
		query = fmt.Sprintf("SELECT %s FROM %s WHERE file_md5 = ? AND document_code = ? AND client_id = ? LIMIT 1", rawDocumentsRows, m.table)
		args = []interface{}{fileMd5, documentCode, clientId}
	}
	var resp RawDocuments
	err := m.conn.QueryRowCtx(ctx, &resp, query, args...)
	switch err {
	case nil:
		return &resp, nil
	case sqlx.ErrNotFound:
		return nil, ErrNotFound
	default:
		return nil, err
	}
}

func (m *customRawDocumentsModel) FindByFileNameAndDocumentCode(ctx context.Context, clientId string, fileName string, documentCode string) (*RawDocuments, error) {
	var query string
	var args []interface{}
	if strings.TrimSpace(clientId) == "" {
		query = fmt.Sprintf("SELECT %s FROM %s WHERE file_name = ? AND document_code = ? LIMIT 1", rawDocumentsRows, m.table)
		args = []interface{}{fileName, documentCode}
	} else {
		query = fmt.Sprintf("SELECT %s FROM %s WHERE file_name = ? AND document_code = ? AND client_id = ? LIMIT 1", rawDocumentsRows, m.table)
		args = []interface{}{fileName, documentCode, clientId}
	}
	var resp RawDocuments
	err := m.conn.QueryRowCtx(ctx, &resp, query, args...)
	switch err {
	case nil:
		return &resp, nil
	case sqlx.ErrNotFound:
		return nil, ErrNotFound
	default:
		return nil, err
	}
}

func (m *customRawDocumentsModel) FindByFileName(ctx context.Context, clientId string, fileName string) (*RawDocuments, error) {
	var query string
	var args []interface{}
	if strings.TrimSpace(clientId) == "" {
		query = fmt.Sprintf("SELECT %s FROM %s WHERE file_name = ? LIMIT 1", rawDocumentsRows, m.table)
		args = []interface{}{fileName}
	} else {
		query = fmt.Sprintf("SELECT %s FROM %s WHERE file_name = ? AND client_id = ? LIMIT 1", rawDocumentsRows, m.table)
		args = []interface{}{fileName, clientId}
	}
	var resp RawDocuments
	err := m.conn.QueryRowCtx(ctx, &resp, query, args...)
	switch err {
	case nil:
		return &resp, nil
	case sqlx.ErrNotFound:
		return nil, ErrNotFound
	default:
		return nil, err
	}
}

func (m *customRawDocumentsModel) FindDistinctTags(ctx context.Context, clientId string, documentCode string) ([]string, error) {
	query := fmt.Sprintf("SELECT DISTINCT tag FROM %s WHERE tag IS NOT NULL AND tag != ''", m.table)
	args := []interface{}{}
	if strings.TrimSpace(clientId) != "" {
		query += " AND client_id = ?"
		args = append(args, clientId)
	}

	if documentCode != "" {
		query += " AND document_code = ?"
		args = append(args, documentCode)
	}

	query += " ORDER BY tag"

	var tags []string
	err := m.conn.QueryRowsCtx(ctx, &tags, query, args...)
	if err != nil && err != sqlx.ErrNotFound {
		return nil, err
	}

	// 过滤空字符串
	var filteredTags []string
	for _, tag := range tags {
		if tag != "" {
			filteredTags = append(filteredTags, tag)
		}
	}

	return filteredTags, nil
}

func (m *customRawDocumentsModel) FindAll(ctx context.Context, clientId string) ([]*RawDocuments, error) {
	query := fmt.Sprintf("SELECT %s FROM %s", rawDocumentsRows, m.table)
	args := []interface{}{}
	if strings.TrimSpace(clientId) != "" {
		query += " WHERE client_id = ?"
		args = append(args, clientId)
	}
	query += " ORDER BY created_at DESC"
	var documents []*RawDocuments
	err := m.conn.QueryRowsCtx(ctx, &documents, query, args...)
	return documents, err
}

// CountAuditedByTag 统计 tag = ? 的记录数（无论是否审核）
func (m *customRawDocumentsModel) CountAuditedByTag(ctx context.Context, clientId string, tag string) (int64, error) {
	if strings.TrimSpace(clientId) == "" {
		query := fmt.Sprintf("SELECT COUNT(1) FROM %s WHERE tag = ?", m.table)
		var count int64
		err := m.conn.QueryRowCtx(ctx, &count, query, tag)
		return count, err
	}
	query := fmt.Sprintf("SELECT COUNT(1) FROM %s WHERE tag = ? AND client_id = ?", m.table)
	var count int64
	err := m.conn.QueryRowCtx(ctx, &count, query, tag, clientId)
	return count, err
}

// ClearTag 将 raw_documents 中 tag = ? 的记录的 tag 置为空字符串
func (m *customRawDocumentsModel) ClearTag(ctx context.Context, clientId string, tag string) error {
	if strings.TrimSpace(clientId) == "" {
		query := fmt.Sprintf("UPDATE %s SET tag = '' WHERE tag = ?", m.table)
		_, err := m.conn.ExecCtx(ctx, query, tag)
		return err
	}
	query := fmt.Sprintf("UPDATE %s SET tag = '' WHERE tag = ? AND client_id = ?", m.table)
	_, err := m.conn.ExecCtx(ctx, query, tag, clientId)
	return err
}

func (m *customRawDocumentsModel) execUpdateRows(ctx context.Context, query string, args ...interface{}) (int64, error) {
	ret, err := m.conn.ExecCtx(ctx, query, args...)
	if err != nil {
		return 0, err
	}
	return ret.RowsAffected()
}

func (m *customRawDocumentsModel) UpdateStatusAndMsg(ctx context.Context, clientId string, id int64, status, statusMsg string) (int64, error) {
	if id <= 0 {
		return 0, nil
	}
	if strings.TrimSpace(clientId) == "" {
		query := fmt.Sprintf("UPDATE %s SET `status` = ?, `status_msg` = ?, `updated_at` = ? WHERE `id` = ?", m.table)
		return m.execUpdateRows(ctx, query, status, statusMsg, time.Now(), id)
	}
	query := fmt.Sprintf("UPDATE %s SET `status` = ?, `status_msg` = ?, `updated_at` = ? WHERE `id` = ? AND `client_id` = ?", m.table)
	return m.execUpdateRows(ctx, query, status, statusMsg, time.Now(), id, clientId)
}

func (m *customRawDocumentsModel) UpdateStatusOnly(ctx context.Context, clientId string, id int64, status string) (int64, error) {
	if id <= 0 {
		return 0, nil
	}
	if strings.TrimSpace(clientId) == "" {
		query := fmt.Sprintf("UPDATE %s SET `status` = ?, `updated_at` = ? WHERE `id` = ?", m.table)
		return m.execUpdateRows(ctx, query, status, time.Now(), id)
	}
	query := fmt.Sprintf("UPDATE %s SET `status` = ?, `updated_at` = ? WHERE `id` = ? AND `client_id` = ?", m.table)
	return m.execUpdateRows(ctx, query, status, time.Now(), id, clientId)
}

func (m *customRawDocumentsModel) ClearStatusMsg(ctx context.Context, clientId string, id int64) (int64, error) {
	if id <= 0 {
		return 0, nil
	}
	if strings.TrimSpace(clientId) == "" {
		query := fmt.Sprintf("UPDATE %s SET `status_msg` = '', `updated_at` = ? WHERE `id` = ?", m.table)
		return m.execUpdateRows(ctx, query, time.Now(), id)
	}
	query := fmt.Sprintf("UPDATE %s SET `status_msg` = '', `updated_at` = ? WHERE `id` = ? AND `client_id` = ?", m.table)
	return m.execUpdateRows(ctx, query, time.Now(), id, clientId)
}

func (m *customRawDocumentsModel) ClearAuditFields(ctx context.Context, clientId string, id int64) error {
	if id <= 0 {
		return nil
	}
	if strings.TrimSpace(clientId) == "" {
		query := fmt.Sprintf("UPDATE %s SET `is_audit` = 0, `audit_user` = '', `audited_at` = NULL, `updated_at` = ? WHERE `id` = ?", m.table)
		_, err := m.conn.ExecCtx(ctx, query, time.Now(), id)
		return err
	}
	query := fmt.Sprintf("UPDATE %s SET `is_audit` = 0, `audit_user` = '', `audited_at` = NULL, `updated_at` = ? WHERE `id` = ? AND `client_id` = ?", m.table)
	_, err := m.conn.ExecCtx(ctx, query, time.Now(), id, clientId)
	return err
}

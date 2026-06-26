package model

import (
	"context"
	"strings"
	"time"

	"knowsource/common/constants"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ AsyncTaskModel = (*customAsyncTaskModel)(nil)

type (
	// AsyncTaskModel is an interface to be customized, add more methods here,
	// and implement the added methods in customAsyncTaskModel.
	AsyncTaskModel interface {
		asyncTaskModel
		WithSession(session sqlx.Session) AsyncTaskModel
		FindPendingTasks(ctx context.Context) ([]*AsyncTask, error)
		CountPendingInit(ctx context.Context) (int64, error)
		FindByTaskTypeAndSourceId(ctx context.Context, taskType string, sourceId int64) (*AsyncTask, error)
		FindActiveByTaskTypeAndSourceId(ctx context.Context, clientId string, taskType string, sourceId int64) (*AsyncTask, error)
		FindByClientId(ctx context.Context, clientId string, taskType string, status string, offset, limit int64) ([]*AsyncTask, error)
		CountByClientId(ctx context.Context, clientId string, taskType string, status string) (int64, error)
		UpdateStatus(ctx context.Context, id int64, status string, result string) error
		Create(ctx context.Context, taskType, taskDesc string, sourceId int64, sourceKey string) error
		CreateWithClientId(ctx context.Context, clientId string, taskType, taskDesc string, sourceId int64, sourceKey string) (int64, error)
		ClaimRunning(ctx context.Context, id int64) (bool, error)
		CancelActiveByTaskTypeAndSourceId(ctx context.Context, clientId string, taskType string, sourceId int64, result string) (bool, error)
		// CancelActiveByTaskTypeAndSourceIdAnyClient 不限制 client_id（兼容历史任务 client_id 为空）
		CancelActiveByTaskTypeAndSourceIdAnyClient(ctx context.Context, taskType string, sourceId int64, result string) (bool, error)
		// CancelActiveById 取消当前租户的 init/running 任务
		CancelActiveById(ctx context.Context, clientId string, id int64, result string) (bool, error)
		// FindRunningTimeoutTasks 查询已超时的 running 任务
		FindRunningTimeoutTasks(ctx context.Context, timeoutBefore time.Time, limit int64) ([]*AsyncTask, error)
		// CancelRunningById 将 running 任务按 id 置为 canceled（用于超时回收）
		CancelRunningById(ctx context.Context, id int64, result string) (bool, error)
	}

	customAsyncTaskModel struct {
		*defaultAsyncTaskModel
	}
)

// NewAsyncTaskModel returns a model for the database table.
func NewAsyncTaskModel(conn sqlx.SqlConn) AsyncTaskModel {
	return &customAsyncTaskModel{
		defaultAsyncTaskModel: newAsyncTaskModel(conn),
	}
}

func (m *customAsyncTaskModel) WithSession(session sqlx.Session) AsyncTaskModel {
	return NewAsyncTaskModel(sqlx.NewSqlConnFromSession(session))
}

// FindPendingTasks 查询状态为未开始(1)的任务
func (m *customAsyncTaskModel) FindPendingTasks(ctx context.Context) ([]*AsyncTask, error) {
	query := `select ` + asyncTaskRows + ` from ` + m.table + ` where status = ? order by created_at asc`
	var resp []*AsyncTask
	err := m.conn.QueryRowsCtx(ctx, &resp, query, constants.AsyncTaskStatusInit)
	return resp, err
}

// CountPendingInit 统计 status=init 的任务数（用于同步 Redis 唤醒 key）
func (m *customAsyncTaskModel) CountPendingInit(ctx context.Context) (int64, error) {
	query := `select count(1) from ` + m.table + ` where status = ?`
	var cnt int64
	err := m.conn.QueryRowCtx(ctx, &cnt, query, constants.AsyncTaskStatusInit)
	return cnt, err
}

// FindByTaskTypeAndSourceId 根据任务类型和源ID查询任务
func (m *customAsyncTaskModel) FindByTaskTypeAndSourceId(ctx context.Context, taskType string, sourceId int64) (*AsyncTask, error) {
	query := `select ` + asyncTaskRows + ` from ` + m.table + ` where task_type = ? and source_id = ? limit 1`
	var resp AsyncTask
	err := m.conn.QueryRowCtx(ctx, &resp, query, taskType, sourceId)
	switch err {
	case nil:
		return &resp, nil
	case sqlx.ErrNotFound:
		return nil, nil
	default:
		return nil, err
	}
}

// FindActiveByTaskTypeAndSourceId 查询 pending/running 的任务（用于幂等与“正在执行”判断）
func (m *customAsyncTaskModel) FindActiveByTaskTypeAndSourceId(ctx context.Context, clientId string, taskType string, sourceId int64) (*AsyncTask, error) {
	query := `select ` + asyncTaskRows + ` from ` + m.table + ` where client_id = ? and task_type = ? and source_id = ? and status in (?, ?) order by updated_at desc, created_at desc limit 1`
	var resp AsyncTask
	err := m.conn.QueryRowCtx(ctx, &resp, query, clientId, taskType, sourceId, constants.AsyncTaskStatusInit, constants.AsyncTaskStatusRunning)
	switch err {
	case nil:
		return &resp, nil
	case sqlx.ErrNotFound:
		return nil, nil
	default:
		return nil, err
	}
}

func (m *customAsyncTaskModel) FindByClientId(ctx context.Context, clientId string, taskType string, status string, offset, limit int64) ([]*AsyncTask, error) {
	query := `select ` + asyncTaskRows + ` from ` + m.table + ` where client_id = ?`
	args := []interface{}{clientId}
	if strings.TrimSpace(taskType) != "" {
		query += ` and task_type = ?`
		args = append(args, taskType)
	}
	if strings.TrimSpace(status) != "" {
		query += ` and status = ?`
		args = append(args, strings.TrimSpace(status))
	}
	query += ` order by created_at desc limit ? offset ?`
	args = append(args, limit, offset)

	var resp []*AsyncTask
	err := m.conn.QueryRowsCtx(ctx, &resp, query, args...)
	return resp, err
}

func (m *customAsyncTaskModel) CountByClientId(ctx context.Context, clientId string, taskType string, status string) (int64, error) {
	query := `select count(1) from ` + m.table + ` where client_id = ?`
	args := []interface{}{clientId}
	if strings.TrimSpace(taskType) != "" {
		query += ` and task_type = ?`
		args = append(args, taskType)
	}
	if strings.TrimSpace(status) != "" {
		query += ` and status = ?`
		args = append(args, strings.TrimSpace(status))
	}

	var cnt int64
	err := m.conn.QueryRowCtx(ctx, &cnt, query, args...)
	return cnt, err
}

// UpdateStatus 更新任务状态
func (m *customAsyncTaskModel) UpdateStatus(ctx context.Context, id int64, status string, result string) error {
	query := `update ` + m.table + ` set status = ?, updated_at = now(),execute_result = ? where id = ?`
	_, err := m.conn.ExecCtx(ctx, query, status, result, id)
	return err
}

func (m *customAsyncTaskModel) Create(ctx context.Context, taskType, taskDesc string, sourceId int64, sourceKey string) error {
	clientId, _ := ctx.Value("clientId").(string)
	clientId = strings.TrimSpace(clientId)
	task := AsyncTask{
		ClientId:  clientId,
		TaskType:  taskType,
		TaskDesc:  taskDesc,
		SourceId:  sourceId,
		SourceKey: sourceKey,
		Status:    constants.AsyncTaskStatusInit,
	}
	_, err := m.Insert(ctx, &task)
	return err
}

func (m *customAsyncTaskModel) CreateWithClientId(ctx context.Context, clientId string, taskType, taskDesc string, sourceId int64, sourceKey string) (int64, error) {
	task := AsyncTask{
		ClientId:  clientId,
		TaskType:  taskType,
		TaskDesc:  taskDesc,
		SourceId:  sourceId,
		SourceKey: sourceKey,
		Status:    constants.AsyncTaskStatusInit,
	}
	res, err := m.Insert(ctx, &task)
	if err != nil {
		return 0, err
	}
	id, _ := res.LastInsertId()
	return id, nil
}

// ClaimRunning 将 init 的任务原子置为 running，避免并发重复执行
func (m *customAsyncTaskModel) ClaimRunning(ctx context.Context, id int64) (bool, error) {
	query := `update ` + m.table + ` set status = ?, updated_at = now() where id = ? and status = ?`
	ret, err := m.conn.ExecCtx(ctx, query, constants.AsyncTaskStatusRunning, id, constants.AsyncTaskStatusInit)
	if err != nil {
		return false, err
	}
	ra, err := ret.RowsAffected()
	if err != nil {
		return false, err
	}
	return ra > 0, nil
}

// CancelActiveByTaskTypeAndSourceId 取消 init/running 的任务
func (m *customAsyncTaskModel) CancelActiveByTaskTypeAndSourceId(ctx context.Context, clientId string, taskType string, sourceId int64, result string) (bool, error) {
	query := `update ` + m.table + ` set status = ?, updated_at = now(), execute_result = ? where client_id = ? and task_type = ? and source_id = ? and status in (?, ?)`
	ret, err := m.conn.ExecCtx(ctx, query, constants.AsyncTaskStatusCanceled, result, clientId, taskType, sourceId, constants.AsyncTaskStatusInit, constants.AsyncTaskStatusRunning)
	if err != nil {
		return false, err
	}
	ra, err := ret.RowsAffected()
	if err != nil {
		return false, err
	}
	return ra > 0, nil
}

// CancelActiveByTaskTypeAndSourceIdAnyClient 取消 init/running 任务（不校验 client_id，兼容历史数据）
func (m *customAsyncTaskModel) CancelActiveByTaskTypeAndSourceIdAnyClient(ctx context.Context, taskType string, sourceId int64, result string) (bool, error) {
	query := `update ` + m.table + ` set status = ?, updated_at = now(), execute_result = ? where task_type = ? and source_id = ? and status in (?, ?)`
	ret, err := m.conn.ExecCtx(ctx, query, constants.AsyncTaskStatusCanceled, result, taskType, sourceId, constants.AsyncTaskStatusInit, constants.AsyncTaskStatusRunning)
	if err != nil {
		return false, err
	}
	ra, err := ret.RowsAffected()
	if err != nil {
		return false, err
	}
	return ra > 0, nil
}

// CancelActiveById 取消当前租户的 init/running 任务
func (m *customAsyncTaskModel) CancelActiveById(ctx context.Context, clientId string, id int64, result string) (bool, error) {
	query := `update ` + m.table + ` set status = ?, updated_at = now(), execute_result = ? where id = ? and client_id = ? and status in (?, ?)`
	ret, err := m.conn.ExecCtx(ctx, query, constants.AsyncTaskStatusCanceled, result, id, clientId, constants.AsyncTaskStatusInit, constants.AsyncTaskStatusRunning)
	if err != nil {
		return false, err
	}
	ra, err := ret.RowsAffected()
	if err != nil {
		return false, err
	}
	return ra > 0, nil
}

// FindRunningTimeoutTasks 查询已超时的 running 任务
func (m *customAsyncTaskModel) FindRunningTimeoutTasks(ctx context.Context, timeoutBefore time.Time, limit int64) ([]*AsyncTask, error) {
	if limit <= 0 {
		limit = 100
	}
	query := `select ` + asyncTaskRows + ` from ` + m.table + ` where status = ? and updated_at < ? order by updated_at asc limit ?`
	var resp []*AsyncTask
	err := m.conn.QueryRowsCtx(ctx, &resp, query, constants.AsyncTaskStatusRunning, timeoutBefore, limit)
	return resp, err
}

// CancelRunningById 将 running 任务按 id 置为 canceled（用于超时回收）
func (m *customAsyncTaskModel) CancelRunningById(ctx context.Context, id int64, result string) (bool, error) {
	query := `update ` + m.table + ` set status = ?, updated_at = now(), execute_result = ? where id = ? and status = ?`
	ret, err := m.conn.ExecCtx(ctx, query, constants.AsyncTaskStatusCanceled, result, id, constants.AsyncTaskStatusRunning)
	if err != nil {
		return false, err
	}
	ra, err := ret.RowsAffected()
	if err != nil {
		return false, err
	}
	return ra > 0, nil
}

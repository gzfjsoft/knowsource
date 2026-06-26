package executor

import (
	"context"
	"errors"
	"fmt"
	"knowsource/api/internal/svc"
	knowsourceLogic "knowsource/api/internal/logic/knowsource"
	"strings"
	"time"

	"knowsource/common/asynctask"
	"knowsource/common/asynctasksignal"
	"knowsource/common/constants"
	"knowsource/model"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/redis"
)

const defaultRedisPoll = 300 * time.Millisecond
const defaultTaskTimeout = 30 * time.Minute
const defaultTimeoutSweepInterval = 30 * time.Second

type TaskProcessor struct {
	svcCtx               *svc.ServiceContext
	taskModel            model.AsyncTaskModel
	redis                *redis.Redis
	redisPoll            time.Duration // pending key 不存在时的 Redis 轮询间隔（不轮询 MySQL）
	taskTimeout          time.Duration // 单任务执行超时时间（超时自动中断）
	timeoutSweepInterval time.Duration // 超时 running 任务回收间隔
	lastTimeoutSweepAt   time.Time
}

type TaskExecutor interface {
	Execute(ctx context.Context, task *model.AsyncTask) (string, error)
}

type TaskExecutorMap map[string]TaskExecutor

func NewTaskProcessor(svcCtx *svc.ServiceContext) *TaskProcessor {
	return &TaskProcessor{
		svcCtx:               svcCtx,
		taskModel:            model.NewAsyncTaskModel(svcCtx.Mysql),
		redis:                svcCtx.RedisClient,
		redisPoll:            defaultRedisPoll,
		taskTimeout:          defaultTaskTimeout,
		timeoutSweepInterval: defaultTimeoutSweepInterval,
	}
}

// WithRedisPollInterval 设置仅在「无 pending Redis key」时的休眠间隔；有任务时由 key 唤醒，不依赖该间隔轮询表。
func (tp *TaskProcessor) WithRedisPollInterval(d time.Duration) *TaskProcessor {
	if d > 0 {
		tp.redisPoll = d
	}
	return tp
}

// WithTaskTimeout 设置单任务超时时间（<=0 时保留默认值）。
func (tp *TaskProcessor) WithTaskTimeout(d time.Duration) *TaskProcessor {
	if d > 0 {
		tp.taskTimeout = d
	}
	return tp
}

func (tp *TaskProcessor) Start(executors TaskExecutorMap) {
	if tp.redis == nil {
		logx.Error("Async task processor: RedisClient is nil, processor not started")
		return
	}
	if tp.redisPoll <= 0 {
		tp.redisPoll = defaultRedisPoll
	}
	if tp.taskTimeout <= 0 {
		tp.taskTimeout = defaultTaskTimeout
	}
	if tp.timeoutSweepInterval <= 0 {
		tp.timeoutSweepInterval = defaultTimeoutSweepInterval
	}

	ctx := context.Background()
	tp.bootstrapPendingKey(ctx)

	go func() {
		for {
			tp.waitAndProcessRound(ctx, executors)
		}
	}()

	logx.Infof("Async task processor started (redis key %s, idle poll %v, task timeout %v)", asynctasksignal.KeyPending, tp.redisPoll, tp.taskTimeout)
}

func (tp *TaskProcessor) bootstrapPendingKey(ctx context.Context) {
	n, err := tp.taskModel.CountPendingInit(ctx)
	if err != nil {
		logx.Errorf("async_task bootstrap CountPendingInit: %v", err)
		return
	}
	if n > 0 {
		asynctasksignal.SetPendingIf(ctx, tp.redis, true)
		logx.Infof("async_task bootstrap: %d init row(s), pending key set", n)
	}
}

func (tp *TaskProcessor) syncPendingKey(ctx context.Context) {
	n, err := tp.taskModel.CountPendingInit(ctx)
	if err != nil {
		logx.Errorf("async_task CountPendingInit for redis sync: %v", err)
		return
	}
	asynctasksignal.SetPendingIf(ctx, tp.redis, n > 0)
}

// waitAndProcessRound：只轮询 Redis；发现 pending key 后加 inspect 锁再查表；无 init 任务则删 key 并解锁。
func (tp *TaskProcessor) waitAndProcessRound(ctx context.Context, executors TaskExecutorMap) {
	tp.sweepTimeoutRunning(ctx)

	exists, err := tp.redis.ExistsCtx(ctx, asynctasksignal.KeyPending)
	if err != nil {
		logx.Errorf("async_task redis Exists: %v", err)
		time.Sleep(tp.redisPoll)
		return
	}
	if !exists {
		time.Sleep(tp.redisPoll)
		return
	}

	locked, err := tp.redis.SetnxExCtx(ctx, asynctasksignal.KeyInspectLock, "1", asynctasksignal.InspectLockSeconds)
	if err != nil {
		logx.Errorf("async_task inspect lock: %v", err)
		time.Sleep(tp.redisPoll)
		return
	}
	if !locked {
		time.Sleep(50 * time.Millisecond)
		return
	}

	tasks, err := tp.taskModel.FindPendingTasks(ctx)
	if err != nil {
		logx.Errorf("Failed to fetch pending tasks: %v", err)
		_, _ = tp.redis.DelCtx(ctx, asynctasksignal.KeyInspectLock)
		time.Sleep(tp.redisPoll)
		return
	}
	if len(tasks) == 0 {
		_, _ = tp.redis.DelCtx(ctx, asynctasksignal.KeyPending)
		_, _ = tp.redis.DelCtx(ctx, asynctasksignal.KeyInspectLock)
		time.Sleep(tp.redisPoll)
		return
	}

	_, _ = tp.redis.DelCtx(ctx, asynctasksignal.KeyInspectLock)

	tp.processTasksBatch(ctx, executors)
	tp.syncPendingKey(ctx)
}

func (tp *TaskProcessor) sweepTimeoutRunning(ctx context.Context) {
	if tp.taskTimeout <= 0 {
		return
	}
	now := time.Now()
	if !tp.lastTimeoutSweepAt.IsZero() && now.Sub(tp.lastTimeoutSweepAt) < tp.timeoutSweepInterval {
		return
	}
	tp.lastTimeoutSweepAt = now

	timeoutBefore := now.Add(-tp.taskTimeout)
	tasks, err := tp.taskModel.FindRunningTimeoutTasks(ctx, timeoutBefore, 200)
	if err != nil {
		logx.Errorf("async_task timeout sweep query failed: %v", err)
		return
	}
	if len(tasks) == 0 {
		return
	}

	for _, task := range tasks {
		if task == nil {
			continue
		}
		result := fmt.Sprintf("任务执行超时（>%v），系统自动中断", tp.taskTimeout)
		ok, cancelErr := tp.taskModel.CancelRunningById(ctx, task.Id, result)
		if cancelErr != nil {
			logx.Errorf("async_task timeout sweep cancel failed, task=%d err=%v", task.Id, cancelErr)
			continue
		}
		if !ok {
			continue
		}
		knowsourceLogic.RevertAfterAsyncTaskCanceled(ctx, tp.svcCtx, task)
		asynctasksignal.BumpClientWatermark(ctx, tp.redis, task.ClientId)
		logx.Infof("async_task timeout sweep canceled task=%d type=%s sourceId=%d", task.Id, task.TaskType, task.SourceId)
	}
}

func (tp *TaskProcessor) processTasksBatch(ctx context.Context, executors TaskExecutorMap) {
	startTime := time.Now()
	logx.Infof("Starting task batch processing at %v", startTime)

	tasks, err := tp.taskModel.FindPendingTasks(ctx)
	if err != nil {
		logx.Errorf("Failed to fetch pending tasks: %v", err)
		return
	}

	if len(tasks) == 0 {
		logx.Info("No pending tasks found")
		return
	}

	logx.Infof("Found %d pending tasks", len(tasks))

	taskMap := tp.deduplicateTasks(ctx, tasks)
	logx.Infof("After deduplication: %d unique tasks to process", len(taskMap))

	processedCount := 0
	successCount := 0
	failedCount := 0

	for _, task := range taskMap {
		if tp.executeTask(ctx, task, executors) {
			successCount++
		} else {
			failedCount++
		}
		processedCount++
	}

	duration := time.Since(startTime)
	logx.Infof("Task batch processing completed in %v. Processed: %d, Success: %d, Failed: %d",
		duration, processedCount, successCount, failedCount)
}

func (tp *TaskProcessor) deduplicateTasks(ctx context.Context, tasks []*model.AsyncTask) map[string]*model.AsyncTask {
	taskMap := make(map[string]*model.AsyncTask)

	for _, task := range tasks {
		key := fmt.Sprintf("%s_%d", task.TaskType, task.SourceId)

		if existingTask, exists := taskMap[key]; exists {
			if task.CreatedAt.After(existingTask.CreatedAt) {
				taskMap[key] = task
				if uerr := tp.taskModel.UpdateStatus(ctx, existingTask.Id, constants.AsyncTaskStatusCanceled, "重复任务，取消旧任务"); uerr == nil {
					asynctasksignal.BumpClientWatermark(ctx, tp.redis, existingTask.ClientId)
				}
			}
		} else {
			taskMap[key] = task
		}
	}

	return taskMap
}

func (tp *TaskProcessor) executeTask(ctx context.Context, task *model.AsyncTask, executors TaskExecutorMap) bool {
	taskKey := fmt.Sprintf("%s_%d", task.TaskType, task.SourceId)
	startTime := time.Now()

	logx.Infof("Executing task [%s] ID: %d, Type: %s, SourceId: %d",
		taskKey, task.Id, task.TaskType, task.SourceId)

	claimed, claimErr := tp.taskModel.ClaimRunning(ctx, task.Id)
	if claimErr != nil {
		logx.Errorf("Failed to claim task %d: %v", task.Id, claimErr)
		return false
	}
	if !claimed {
		logx.Infof("Task %d already claimed/updated, skip", task.Id)
		return true
	}
	asynctasksignal.BumpClientWatermark(ctx, tp.redis, task.ClientId)

	executor, exists := executors[task.TaskType]
	if !exists {
		logx.Errorf("No executor found for task type: %s", task.TaskType)
		tp.updateTaskStatus(ctx, task.ClientId, task.Id, constants.AsyncTaskStatusFailed, "No executor found")
		return false
	}

	runCtx := ctx
	var cancel context.CancelFunc
	if tp.taskTimeout > 0 {
		runCtx, cancel = context.WithTimeout(ctx, tp.taskTimeout)
		defer cancel()
	}

	res, err := executor.Execute(runCtx, task)
	duration := time.Since(startTime)

	if runCtx.Err() == context.DeadlineExceeded {
		msg := fmt.Sprintf("任务执行超时（>%v），系统自动中断", tp.taskTimeout)
		if res != "" {
			msg = strings.TrimSpace(msg + "; " + res)
		}
		logx.Errorf("Task [%s] timed out in %v", taskKey, duration)
		knowsourceLogic.RevertAfterAsyncTaskCanceled(ctx, tp.svcCtx, task)
		tp.updateTaskStatus(ctx, task.ClientId, task.Id, constants.AsyncTaskStatusCanceled, msg)
		return false
	}

	if err != nil {
		if errors.Is(err, asynctask.ErrTaskCancelled) {
			logx.Infof("Task [%s] cancelled in %v", taskKey, duration)
			tp.updateTaskStatus(ctx, task.ClientId, task.Id, constants.AsyncTaskStatusCanceled, res)
			return true
		}
		logx.Errorf("Task [%s] execution failed in %v: %v", taskKey, duration, err)
		tp.updateTaskStatus(ctx, task.ClientId, task.Id, constants.AsyncTaskStatusFailed, res)
		return false
	}

	logx.Infof("Task [%s] executed successfully in %v", taskKey, duration)
	tp.updateTaskStatus(ctx, task.ClientId, task.Id, constants.AsyncTaskStatusSuccess, res)
	return true
}

func (tp *TaskProcessor) updateTaskStatus(ctx context.Context, clientId string, taskId int64, status string, result string) {
	err := tp.taskModel.UpdateStatus(ctx, taskId, status, result)
	if err != nil {
		logx.Errorf("Failed to update task %d status to %s: %v", taskId, status, err)
	} else {
		logx.Infof("Updated task %d status to %s", taskId, status)
		asynctasksignal.BumpClientWatermark(ctx, tp.redis, clientId)
	}
}

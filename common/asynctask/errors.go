package asynctask

import "errors"

// ErrTaskCancelled 异步任务被用户或系统取消
var ErrTaskCancelled = errors.New("task_cancelled")

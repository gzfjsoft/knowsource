package constants

// Async task row status in `async_task.status` (varchar).
const (
	AsyncTaskStatusInit     = "init"
	AsyncTaskStatusRunning  = "running"
	AsyncTaskStatusCanceled = "canceled"
	AsyncTaskStatusSuccess  = "success"
	AsyncTaskStatusFailed   = "failed"
)

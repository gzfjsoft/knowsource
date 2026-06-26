package bootstrap

// ServerStartedInInitMode 为 true 表示进程因 MySQL/Redis 未就绪而以「仅初始化」方式启动，需重启后才能提供完整 API。
var ServerStartedInInitMode bool

package asynctasksignal

import (
	"context"
	"strconv"
	"strings"
	"time"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/redis"
)

// Redis keys：用独立 key 唤醒 worker，避免高频轮询 MySQL。
const (
	KeyPending      = "knowsource:async_task:pending"
	KeyProducerLock = "knowsource:async_task:producer_lock"
	KeyInspectLock  = "knowsource:async_task:inspect_lock"
	pendingTTL      = 86400 // 秒，防止 worker 异常后 key 永久残留
	producerLockTTL = 3
	// InspectLockSeconds worker 查表前持有的 inspect 锁 TTL（秒）
	InspectLockSeconds = 15
	lockSpinDeadline   = 2 * time.Second
	lockSpinInterval   = 5 * time.Millisecond
	watermarkTTL       = 86400 * 30 // 秒，每次 bump 会刷新 TTL
)

// ClientWatermarkKey 每个租户一条，值为 Unix 毫秒时间戳字符串；async_task 表有变更时 bump，供队列页轻量轮询。
func ClientWatermarkKey(clientId string) string {
	return "knowsource:async_task:wm:" + strings.TrimSpace(clientId)
}

// BumpClientWatermark 在 async_task 对该租户有插入/状态更新后调用，刷新「最后变更」时间。
func BumpClientWatermark(ctx context.Context, r *redis.Redis, clientId string) {
	if r == nil {
		return
	}
	clientId = strings.TrimSpace(clientId)
	if clientId == "" {
		return
	}
	// 使用纳秒时间戳，避免同一毫秒内多次 bump 导致前端比对不变
	ns := strconv.FormatInt(time.Now().UnixNano(), 10)
	if err := r.SetexCtx(ctx, ClientWatermarkKey(clientId), ns, watermarkTTL); err != nil {
		logx.WithContext(ctx).Errorf("async_task bump watermark: %v", err)
	}
}

// NotifyPending 在成功写入 async_task 后调用：加锁 → 设置 pending key → 释放锁；并 bump 当前租户 watermark。
func NotifyPending(ctx context.Context, r *redis.Redis, clientId string) error {
	if r == nil {
		return nil
	}
	deadline := time.Now().Add(lockSpinDeadline)
	for time.Now().Before(deadline) {
		ok, err := r.SetnxExCtx(ctx, KeyProducerLock, "1", producerLockTTL)
		if err != nil {
			logx.WithContext(ctx).Errorf("async_task signal producer lock: %v", err)
			return err
		}
		if ok {
			if err := r.SetexCtx(ctx, KeyPending, "1", pendingTTL); err != nil {
				_, _ = r.DelCtx(ctx, KeyProducerLock)
				logx.WithContext(ctx).Errorf("async_task signal set pending: %v", err)
				return err
			}
			if _, err := r.DelCtx(ctx, KeyProducerLock); err != nil {
				logx.WithContext(ctx).Errorf("async_task signal del producer lock: %v", err)
			}
			BumpClientWatermark(ctx, r, clientId)
			return nil
		}
		time.Sleep(lockSpinInterval)
	}
	// 长时间拿不到锁仍设置 pending，避免丢唤醒
	if err := r.SetexCtx(ctx, KeyPending, "1", pendingTTL); err != nil {
		logx.WithContext(ctx).Errorf("async_task signal set pending (fallback): %v", err)
		return err
	}
	BumpClientWatermark(ctx, r, clientId)
	return nil
}

// SetPendingIf 根据是否仍有 init 任务同步 Redis：有则 SET，无则 DEL。
func SetPendingIf(ctx context.Context, r *redis.Redis, hasInitPending bool) {
	if r == nil {
		return
	}
	if hasInitPending {
		if err := r.SetexCtx(ctx, KeyPending, "1", pendingTTL); err != nil {
			logx.WithContext(ctx).Errorf("async_task sync pending set: %v", err)
		}
		return
	}
	if _, err := r.DelCtx(ctx, KeyPending); err != nil {
		logx.WithContext(ctx).Errorf("async_task sync pending del: %v", err)
	}
}

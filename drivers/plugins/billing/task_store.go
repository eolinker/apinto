// Package billing 任务存储：Redis 优先 + 本地内存回退的双层任务执行器。
package billing

import (
	"context"
	"log"
	"time"

	"github.com/eolinker/apinto/pricing"
	"github.com/eolinker/apinto/resources"
	scope_manager "github.com/eolinker/apinto/scope-manager"
)

// RedisTaskExecutor 双层任务存储：
//   - 主层 Redis：保证多实例间共享，提交时 SetNX 确保幂等
//   - 备层本地内存：Redis 不可用时无缝降级，单实例仍可工作
type RedisTaskExecutor struct {
	cache scope_manager.IProxyOutput[resources.ICache]
	local *pricing.LocalTaskExecutor
	ttl   time.Duration
	ctx   context.Context
}

// NewRedisTaskExecutor 创建双层任务执行器；ttl<=0 走默认值。
func NewRedisTaskExecutor(cache scope_manager.IProxyOutput[resources.ICache], ttl time.Duration) *RedisTaskExecutor {
	if ttl <= 0 {
		ttl = pricing.DefaultTaskTTL
	}
	return &RedisTaskExecutor{
		cache: cache,
		local: pricing.NewLocalTaskExecutor(ttl),
		ttl:   ttl,
		ctx:   context.Background(),
	}
}

func (r *RedisTaskExecutor) getCache() resources.ICache {
	if r.cache == nil {
		return nil
	}
	list := r.cache.List()
	if len(list) < 1 {
		return nil
	}
	for _, c := range list {
		if c != nil {
			return c
		}
	}
	return nil
}

// Submit 提交任务：双写本地 + Redis SetNX；Redis 失败仅记录日志，不中断主链路。
func (r *RedisTaskExecutor) Submit(taskID, provider, resource string, phase pricing.PricingPhase, preResult *pricing.CalculateResult, fields map[string]interface{}) error {
	r.local.Submit(taskID, provider, resource, phase, preResult, fields)

	cache := r.getCache()
	if cache == nil {
		log.Printf("[billing/task] redis unavailable, task stored in local only: task_id=%s", taskID)
		return nil
	}

	entry := &pricing.TaskEntry{
		TaskID:    taskID,
		Provider:  provider,
		Resource:  resource,
		Phase:     phase,
		PreResult: preResult,
		Fields:    fields,
		Status:    pricing.TaskStatusPending,
		CreatedAt: time.Now(),
	}
	data, err := pricing.MarshalTaskEntry(entry)
	if err != nil {
		log.Printf("[billing/task] marshal error: task_id=%s err=%v", taskID, err)
		return err
	}

	ok, setErr := cache.SetNX(r.ctx, pricing.TaskCacheKey(taskID), data, r.ttl).Result()
	if setErr != nil {
		log.Printf("[billing/task] redis SetNX error: task_id=%s err=%v", taskID, setErr)
		return nil
	}
	if !ok {
		log.Printf("[billing/task] task already exists in redis: task_id=%s", taskID)
		return nil
	}
	log.Printf("[billing/task] submitted task to redis: task_id=%s provider=%s resource=%s phase=%s", taskID, provider, resource, phase)
	return nil
}

// Get 优先 Redis、失败回退本地；同时检查已计费标记同步状态。
func (r *RedisTaskExecutor) Get(taskID string) (*pricing.TaskEntry, bool) {
	cache := r.getCache()
	if cache != nil {
		data, err := cache.Get(r.ctx, pricing.TaskCacheKey(taskID)).Result()
		if err == nil && data != "" {
			entry, uerr := pricing.UnmarshalTaskEntry([]byte(data))
			if uerr == nil {
				billedData, berr := cache.Get(r.ctx, pricing.TaskBilledKey(taskID)).Result()
				if berr == nil && billedData != "" {
					entry.Status = pricing.TaskStatusBilled
				}
				return entry, true
			}
		}
	}
	return r.local.Get(taskID)
}

// IsBilled 优先 Redis 查询计费标记键。
func (r *RedisTaskExecutor) IsBilled(taskID string) bool {
	cache := r.getCache()
	if cache != nil {
		_, err := cache.Get(r.ctx, pricing.TaskBilledKey(taskID)).Result()
		if err == nil {
			return true
		}
	}
	return r.local.IsBilled(taskID)
}

// MarkBilled 双写计费标记；Redis SetNX 返回 false 表示其他实例已抢先标记。
func (r *RedisTaskExecutor) MarkBilled(taskID string) bool {
	r.local.MarkBilled(taskID)

	cache := r.getCache()
	if cache == nil {
		return r.local.IsBilled(taskID)
	}
	ok, err := cache.SetNX(r.ctx, pricing.TaskBilledKey(taskID), []byte("1"), r.ttl).Result()
	if err != nil {
		log.Printf("[billing/task] redis SetNX billed error: task_id=%s err=%v", taskID, err)
		return r.local.IsBilled(taskID)
	}
	if !ok {
		log.Printf("[billing/task] task already billed in redis: task_id=%s", taskID)
		return false
	}
	log.Printf("[billing/task] marked task billed in redis: task_id=%s", taskID)
	return true
}

// MarkFailed 标记失败；Redis 中无任务记录时回退到本地标记。
func (r *RedisTaskExecutor) MarkFailed(taskID string) bool {
	r.local.MarkFailed(taskID)

	cache := r.getCache()
	if cache == nil {
		return true
	}
	_, err := cache.Get(r.ctx, pricing.TaskCacheKey(taskID)).Result()
	if err != nil {
		return r.local.MarkFailed(taskID)
	}
	return true
}

// Remove 双删任务条目和计费标记。
func (r *RedisTaskExecutor) Remove(taskID string) {
	r.local.Remove(taskID)

	cache := r.getCache()
	if cache == nil {
		return
	}
	cache.Del(r.ctx, pricing.TaskCacheKey(taskID), pricing.TaskBilledKey(taskID))
}

// Stop 停止本地清理协程；Redis 客户端由 scope_manager 管理。
func (r *RedisTaskExecutor) Stop() {
	r.local.Stop()
}

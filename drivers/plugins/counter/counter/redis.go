package counter

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/go-redis/redis/v8"
)

var _ ICounter = (*RedisCounter)(nil)

type RedisCounter struct {
	ctx   context.Context
	key   string
	redis redis.Cmdable

	client    IClient
	locker    sync.Mutex
	resetTime time.Time

	localCounter ICounter

	lockerKey string
	lockKey   string
	remainKey string
}

func NewRedisCounter(key string, redis redis.Cmdable, client IClient, localCounter ICounter) *RedisCounter {

	return &RedisCounter{
		key:          key,
		redis:        redis,
		client:       client,
		localCounter: localCounter,
		ctx:          context.Background(),
		lockerKey:    fmt.Sprintf("%s:locker", key),
		lockKey:      fmt.Sprintf("%s:lock", key),
		remainKey:    fmt.Sprintf("%s:remain", key),
	}
}

func (r *RedisCounter) Lock(count int64) error {
	r.locker.Lock()
	defer r.locker.Unlock()
	if r.redis == nil {
		// 如果Redis没有配置，使用本地计数器
		return r.localCounter.Lock(count)
	}

	err := r.acquireLock()
	if err != nil {
		if err == redis.ErrClosed {
			return r.localCounter.Lock(count)
		}
		return err
	}

	defer r.releaseLock()

	// 获取最新的次数
	remain, err := r.redis.Get(r.ctx, r.remainKey).Int64()
	if err != nil {
		if err == redis.ErrClosed {
			return r.localCounter.Lock(count)
		} else if err != redis.Nil {
			return err
		}
	}

	remain -= count
	if remain < 0 {
		now := time.Now()
		if now.Sub(r.resetTime) < 10*time.Second {
			return fmt.Errorf("no enough, ddd key:%s, remain:%d, count:%d", r.key, remain+count, count)
		}

		r.resetTime = now
		lock, err := r.redis.Get(r.ctx, r.lockKey).Int64()
		if err != nil {
			if err == redis.ErrClosed {
				return r.localCounter.Lock(count)
			} else if err != redis.Nil {
				return err
			}
		}
		remain, err = getRemainCount(r.client, r.key, count+lock)
		if err != nil {
			return err
		}
	}
	r.redis.Set(r.ctx, r.remainKey, remain, -1)
	r.redis.IncrBy(r.ctx, r.lockKey, count)
	fmt.Println(time.Now().Format("2006-01-02 15:04:05"), "lock", "remain:", remain, "count:", count)
	return nil
}

func (r *RedisCounter) Complete(count int64) error {
	r.locker.Lock()
	defer r.locker.Unlock()
	if r.redis == nil {
		// 如果Redis没有配置，使用本地计数器
		return r.localCounter.Complete(count)
	}

	err := r.acquireLock()
	if err != nil {
		if err == redis.ErrClosed {
			return r.localCounter.Lock(count)
		}
		return err
	}

	defer r.releaseLock()

	r.redis.IncrBy(r.ctx, r.lockKey, -count)
	fmt.Println(time.Now().Format("2006-01-02 15:04:05"), "complete", "count:", count)
	return nil
}

func (r *RedisCounter) RollBack(count int64) error {
	r.locker.Lock()
	defer r.locker.Unlock()
	if r.redis == nil {
		// 如果Redis没有配置，使用本地计数器
		return r.localCounter.RollBack(count)
	}
	err := r.acquireLock()
	if err != nil {
		if err == redis.ErrClosed {
			return r.localCounter.Lock(count)
		}
		return err
	}

	defer r.releaseLock()

	r.redis.IncrBy(r.ctx, r.remainKey, count)
	r.redis.IncrBy(r.ctx, r.lockKey, -count)
	fmt.Println(time.Now().Format("2006-01-02 15:04:05"), "rollback", "count:", count)
	return nil
}

func (r *RedisCounter) ResetClient(client IClient) {
	r.locker.Lock()
	defer r.locker.Unlock()
	r.client = client
}

func (r *RedisCounter) acquireLock() error {
	for {
		// 生成唯一的锁值
		lockValue := time.Now().UnixNano()

		// Redis连接失败，使用本地计数器
		ok, err := r.redis.SetNX(r.ctx, r.lockerKey, lockValue, 10*time.Second).Result()
		if err != nil {
			return err
		}
		if ok {
			// 设置锁成功
			break
		}
	}
	return nil
}

func (r *RedisCounter) releaseLock() error {
	return r.redis.Del(r.ctx, r.lockerKey).Err()
}

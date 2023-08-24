package counter

import (
	"context"
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/eolinker/eosc"

	scope_manager "github.com/eolinker/apinto/scope-manager"

	"github.com/eolinker/apinto/resources"

	"github.com/eolinker/apinto/drivers/counter"

	"github.com/eolinker/eosc/log"
	redis "github.com/go-redis/redis/v8"
)

var _ counter.ICounter = (*RedisCounter)(nil)

type RedisCounter struct {
	ctx   context.Context
	key   string
	redis scope_manager.IProxyOutput[resources.ICache]

	client    scope_manager.IProxyOutput[counter.IClient]
	locker    sync.Mutex
	resetTime time.Time

	localCounter counter.ICounter

	lockerKey string
	lockKey   string
	remainKey string

	variables eosc.Untyped[string, string]
}

func NewRedisCounter(key string, variables eosc.Untyped[string, string], redis scope_manager.IProxyOutput[resources.ICache], client scope_manager.IProxyOutput[counter.IClient]) *RedisCounter {

	return &RedisCounter{
		key:          key,
		redis:        redis,
		client:       client,
		localCounter: NewLocalCounter(key, variables, client),
		ctx:          context.Background(),
		lockerKey:    fmt.Sprintf("%s:locker", key),
		lockKey:      fmt.Sprintf("%s:lock", key),
		remainKey:    fmt.Sprintf("%s:remain", key),
		variables:    variables,
	}
}

func (r *RedisCounter) Lock(count int64) error {
	r.locker.Lock()
	defer r.locker.Unlock()
	list := r.redis.List()
	if len(list) < 1 {
		// Redis不存在，使用本地计数器
		return r.localCounter.Lock(count)
	}
	//if reflect.ValueOf(r.redis).IsNil() {
	//	// 如果Redis没有配置，使用本地计数器
	//	return r.localCounter.Lock(count)
	//}
	var err error
	for _, cache := range list {
		err = r.lock(cache, count)
		if err != nil {
			if err == redis.ErrClosed {
				continue
			}
			return err
		}
		break
	}
	return nil

}

func (r *RedisCounter) lock(cache resources.ICache, count int64) error {
	err := r.acquireLock(cache)
	if err != nil {
		return err
	}

	defer r.releaseLock(cache)

	// 获取最新的次数
	remainCount, err := cache.Get(r.ctx, r.remainKey).Result()
	if err != nil {
		if err != redis.Nil {
			return err
		}
	}
	remain, _ := strconv.ParseInt(remainCount, 10, 64)

	remain -= count
	if remain < 0 {
		now := time.Now()
		if now.Sub(r.resetTime) < 10*time.Second {
			return fmt.Errorf("no enough, ddd key:%s, remain:%d, count:%d", r.key, remain+count, count)
		}

		r.resetTime = now
		var lockCount string
		lockCount, err = cache.Get(r.ctx, r.lockKey).Result()
		if err != nil {
			if err != redis.Nil {
				return err
			}
		}
		lock, _ := strconv.ParseInt(lockCount, 10, 64)
		for _, client := range r.client.List() {
			remain, err = counter.GetRemainCount(client, r.key, count+lock, r.variables)
			if err != nil {
				log.Errorf("get remain count error: %s", err)
				continue
			}
			break
		}
		if err != nil {
			return err
		}
	}
	cache.Set(r.ctx, r.remainKey, []byte(strconv.FormatInt(remain, 10)), -1)
	cache.IncrBy(r.ctx, r.lockKey, count, -1)
	fmt.Println(time.Now().Format("2006-01-02 15:04:05"), "lock", "remain:", remain, "count:", count)
	log.DebugF("lock now: %s,key: %s,remain: %d,count: %d", time.Now().Format("2006-01-02 15:04:05"), r.key, remain, count)
	return nil
}

func (r *RedisCounter) Complete(count int64) error {
	r.locker.Lock()
	defer r.locker.Unlock()
	list := r.redis.List()
	if len(list) < 1 {
		// Redis不存在，使用本地计数器
		return r.localCounter.Lock(count)
	}
	for _, cache := range list {
		err := r.complete(cache, count)
		if err != nil {
			if err == redis.ErrClosed {
				continue
			}
			return err
		}
	}
	return nil
}

func (r *RedisCounter) complete(cache resources.ICache, count int64) error {
	err := r.acquireLock(cache)
	if err != nil {
		return err
	}

	defer r.releaseLock(cache)

	cache.IncrBy(r.ctx, r.lockKey, -count, -1)
	//fmt.Println(time.Now().Format("2006-01-02 15:04:05"), "complete", "count:", count)
	log.DebugF("complete now: %s,key: %s,count: %d", time.Now().Format("2006-01-02 15:04:05"), r.key, count)
	return nil
}

func (r *RedisCounter) RollBack(count int64) error {
	r.locker.Lock()
	defer r.locker.Unlock()
	list := r.redis.List()

	if len(list) < 1 {
		// Redis不存在，使用本地计数器
		return r.localCounter.Lock(count)
	}
	for _, cache := range list {
		err := r.rollback(cache, count)
		if err != nil {
			if err == redis.ErrClosed {
				continue
			}
			return err
		}
	}
	return nil
}

func (r *RedisCounter) rollback(cache resources.ICache, count int64) error {
	err := r.acquireLock(cache)
	if err != nil {
		return err
	}

	defer r.releaseLock(cache)

	cache.IncrBy(r.ctx, r.remainKey, count, -1)
	cache.IncrBy(r.ctx, r.lockKey, -count, -1)
	//fmt.Println(time.Now().Format("2006-01-02 15:04:05"), "rollback", "count:", count)
	log.DebugF("rollback now: %s,key: %s,count: %d", time.Now().Format("2006-01-02 15:04:05"), r.key, count)
	return nil
}

func (r *RedisCounter) acquireLock(cache resources.ICache) error {
	for {
		// 生成唯一的锁值
		lockValue := time.Now().UnixNano()

		// Redis连接失败，使用本地计数器
		ok, err := cache.SetNX(r.ctx, r.lockerKey, []byte(strconv.FormatInt(lockValue, 10)), 10*time.Second).Result()
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

func (r *RedisCounter) releaseLock(cache resources.ICache) error {
	_, err := cache.Del(r.ctx, r.lockerKey).Result()
	return err
}

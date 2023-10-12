package counter

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/eolinker/eosc"
	"github.com/eolinker/eosc/log"

	scope_manager "github.com/eolinker/apinto/scope-manager"

	"github.com/eolinker/apinto/resources"

	"github.com/eolinker/apinto/drivers/counter"

	redis "github.com/go-redis/redis/v8"
)

const (
	redisMode = "redis"
)

var _ counter.ICounter = (*RedisCounter)(nil)

type RedisCounter struct {
	ctx   context.Context
	key   string
	redis scope_manager.IProxyOutput[resources.ICache]

	client        scope_manager.IProxyOutput[counter.IClient]
	counterPusher scope_manager.IProxyOutput[counter.ICountPusher]
	//locker        sync.Mutex
	resetTime time.Time

	localCounter counter.ICounter

	lockerKey string
	lockKey   string
	remainKey string

	variables eosc.Untyped[string, string]
}

func NewRedisCounter(key string, variables eosc.Untyped[string, string], redis scope_manager.IProxyOutput[resources.ICache], client scope_manager.IProxyOutput[counter.IClient], counterPusher scope_manager.IProxyOutput[counter.ICountPusher]) *RedisCounter {

	return &RedisCounter{
		key:           key,
		redis:         redis,
		client:        client,
		localCounter:  NewLocalCounter(key, variables, client, counterPusher),
		counterPusher: counterPusher,
		ctx:           context.Background(),
		lockerKey:     fmt.Sprintf("%s:locker", key),
		lockKey:       fmt.Sprintf("%s:lock", key),
		remainKey:     fmt.Sprintf("%s:remain", key),
		variables:     variables,
	}
}

func (r *RedisCounter) Lock(count int64) error {

	list := r.redis.List()
	if len(list) < 1 {
		// Redis不存在，使用本地计数器
		return r.localCounter.Lock(count)
	}

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
	if err == redis.ErrClosed {
		// 使用本地计数器
		return r.localCounter.RollBack(count)
	}
	return err
}

func (r *RedisCounter) Complete(count int64) error {

	return r.localCounter.Complete(count)
}

func (r *RedisCounter) RollBack(count int64) error {
	list := r.redis.List()

	if len(list) < 1 {
		// Redis不存在，使用本地计数器
		return r.localCounter.RollBack(count)
	}
	var err error
	for _, cache := range list {
		err = r.rollback(cache, count)
		if err != nil {
			if err == redis.ErrClosed {
				continue
			}
			return err
		}
	}
	if err == redis.ErrClosed {
		// 使用本地计数器
		return r.localCounter.RollBack(count)
	}
	return err
}

// lock 次数预扣
func (r *RedisCounter) lock(cache resources.ICache, count int64) error {
	remain, err := cache.DecrBy(r.ctx, r.remainKey, count, -1).Result()
	if err != nil {
		log.Errorf("decr remain error: %s,key: %s", err, r.key)
		return err
	}
	if remain >= 0 {
		// 剩余次数充足，直接返回
		log.DebugF("lock now: %s,key: %s,remain: %d,count: %d", time.Now().Format("2006-01-02 15:04:05"), r.key, remain, count)
		return nil
	}
	// 回滚已经扣的次数
	cache.IncrBy(r.ctx, r.remainKey, count, -1).Result()

	if time.Now().Sub(r.resetTime) < 15*time.Second {
		// 重置时间未到，直接将次数回滚
		return fmt.Errorf("no enough, key:%s, remain:%d, count:%d", r.key, remain+count, count)
	}

	r.resetTime = time.Now()
	err = r.acquireLock(cache)
	if err != nil {
		// 加锁失败，返回报错
		log.Errorf("acquire lock error: %s", err)
		return err
	}

	// 释放分布锁
	defer r.releaseLock(cache)
	// 重新尝试扣减
	remain, err = cache.DecrBy(r.ctx, r.remainKey, count, -1).Result()
	if err != nil {
		log.Errorf("lock decr remain error: %s,key: %s", err, r.key)
		return err
	}
	if remain >= 0 {
		// 当次数大于等于0，此时已经有节点同步过剩余次数，直接返回
		log.DebugF("lock now: %s,key: %s,remain: %d,count: %d", time.Now().Format("2006-01-02 15:04:05"), r.key, remain, count)
		return nil
	}
	// 若此时次数小于0，更新剩余次数
	variables := r.variables.All()
	for _, client := range r.client.List() {
		remain, err = counter.GetRemainCount(client, r.key, count, variables)
		if err != nil {
			log.Errorf("get remain count error: %s", err)
			continue
		}
		break
	}
	if err != nil {
		// 获取次数失败，回滚次数
		cache.IncrBy(r.ctx, r.remainKey, count, -1).Result()
		//return fmt.Errorf("no enough, key:%s, remain:%d, count:%d", r.key, remain+count, count)
		return err
	}
	err = cache.Set(r.ctx, r.remainKey, []byte(strconv.FormatInt(remain, 10)), -1).Result()
	if err != nil {
		return err
	}
	return nil
}

func (r *RedisCounter) rollback(cache resources.ICache, count int64) error {
	remain, err := cache.IncrBy(r.ctx, r.remainKey, count, -1).Result()
	if err != nil {
		log.Errorf("rollback incr remain error: %s,key: %s", err, r.key)
		return err
	}
	log.DebugF("rollback now: %s,key: %s,remain: %d,count: %d", time.Now().Format("2006-01-02 15:04:05"), r.key, remain, count)
	return nil
}

func (r *RedisCounter) acquireLock(cache resources.ICache) error {
	timeoutTicket := time.NewTicker(1 * time.Second)
	for {
		select {
		case <-timeoutTicket.C:
			return fmt.Errorf("acquire lock timeout,key:%s", r.key)
		default:
			lockValue := time.Now().String()
			ok, err := cache.SetNX(r.ctx, r.lockerKey, []byte(lockValue), 10*time.Second).Result()
			if err != nil {
				return err
			}
			if ok {
				return nil
			}
			time.Sleep(100 * time.Millisecond)
		}
	}
}

func (r *RedisCounter) releaseLock(cache resources.ICache) {
	_, err := cache.Del(r.ctx, r.lockerKey).Result()
	if err != nil {
		log.Errorf("release lock error: %s,key: %s", err, r.key)
	}
}

//
//func (r *RedisCounter) complete() error {
//	return nil
//}

//func (r *RedisCounter) lock(cache resources.ICache, count int64) error {
//	// 获取最新的次数
//	remainCount, err := cache.Get(r.ctx, r.remainKey).Result()
//	if err != nil {
//		if err != redis.Nil {
//			return err
//		}
//	}
//	remain, _ := strconv.ParseInt(remainCount, 10, 64)
//
//	remain -= count
//	if remain < 0 {
//		now := time.Now()
//		if now.Sub(r.resetTime) < 10*time.Second {
//			return fmt.Errorf("no enough, key:%s, remain:%d, count:%d", r.key, remain+count, count)
//		}
//
//		r.resetTime = now
//		var lockCount string
//		lockCount, err = cache.Get(r.ctx, r.lockKey).Result()
//		if err != nil {
//			if err != redis.Nil {
//				return err
//			}
//		}
//		lock, _ := strconv.ParseInt(lockCount, 10, 64)
//		variables := r.variables.All()
//		for _, client := range r.client.List() {
//			remain, err = counter.GetRemainCount(client, r.key, count+lock, variables)
//			if err != nil {
//				log.Errorf("get remain count error: %s", err)
//				continue
//			}
//			break
//		}
//		if err != nil {
//			return err
//		}
//	}
//	cache.Set(r.ctx, r.remainKey, []byte(strconv.FormatInt(remain, 10)), -1)
//	cache.IncrBy(r.ctx, r.lockKey, count, -1)
//	log.DebugF("lock now: %s,key: %s,remain: %d,count: %d", time.Now().Format("2006-01-02 15:04:05"), r.key, remain, count)
//	return nil
//}
//
//func (r *RedisCounter) complete(cache resources.ICache, count int64) error {
//
//	cache.IncrBy(r.ctx, r.lockKey, -count, -1)
//	log.DebugF("complete now: %s,key: %s,count: %d", time.Now().Format("2006-01-02 15:04:05"), r.key, count)
//	return nil
//}
//
//func (r *RedisCounter) rollback(cache resources.ICache, count int64) error {
//	cache.IncrBy(r.ctx, r.remainKey, count, -1)
//	cache.IncrBy(r.ctx, r.lockKey, -count, -1)
//	log.DebugF("rollback now: %s,key: %s,count: %d", time.Now().Format("2006-01-02 15:04:05"), r.key, count)
//	return nil
//}

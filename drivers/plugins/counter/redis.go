package counter

import (
	"context"
	"fmt"

	"github.com/eolinker/eosc/log"

	"github.com/eolinker/eosc"

	scope_manager "github.com/eolinker/apinto/scope-manager"

	"github.com/eolinker/apinto/resources"

	redis "github.com/redis/go-redis/v9"
)

type RedisCounter struct {
	ctx       context.Context
	key       string
	redis     scope_manager.IProxyOutput[resources.ICache]
	variables eosc.Untyped[string, string]
}

func NewRedisCounter(key string, redis scope_manager.IProxyOutput[resources.ICache]) *RedisCounter {

	return &RedisCounter{
		key:   key,
		redis: redis,
		ctx:   context.Background(),
	}
}

func (r *RedisCounter) Lock(count int64) error {

	list := r.redis.List()
	if len(list) < 1 {
		// Redis不存在，使用本地计数器
		return nil
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
	return nil
}

func (r *RedisCounter) Complete(count int64) error {
	return nil
}

func (r *RedisCounter) RollBack(count int64) error {
	list := r.redis.List()

	if len(list) < 1 {
		// Redis不存在，使用本地计数器
		return nil
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
	return nil
}

const (
	// key不存在
	keyNotExist = -1

	// 剩余次数不足
	countExceeded = -2
)

// lock 次数预扣
func (r *RedisCounter) lock(cache resources.ICache, count int64) error {
	result, err := cache.Run(r.ctx, lockScript, []string{r.key}, count).Result()
	if err != nil {
		if err == redis.Nil {
			return nil
		}
		return err
	}
	log.DebugF("lock result:%v", result)
	switch t := result.(type) {
	case int64:
		switch int(t) {
		case countExceeded:
			return fmt.Errorf("count exceeded")
		case keyNotExist:
			return fmt.Errorf("key(%s) not exist", r.key)
		}
	}
	return nil
}

func (r *RedisCounter) rollback(cache resources.ICache, count int64) error {
	_, err := cache.Run(r.ctx, callbackScript, []string{r.key}, count).Result()
	if err == redis.Nil {
		return nil
	}
	return err
}

package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/eolinker/eosc/log"

	backoff "github.com/cenkalti/backoff/v4"
	"github.com/eolinker/apinto/resources"
	"github.com/redis/go-redis/v9"
)

func newBoolResult(ok bool, err error) resources.BoolResult {
	return &boolResult{
		ok:  ok,
		err: err,
	}
}

type boolResult struct {
	ok  bool
	err error
}

func (b *boolResult) Result() (bool, error) {
	return b.ok, b.err
}

type statusResult struct {
	err error
}

func (s *statusResult) Result() error {
	return s.err
}

type CmdAble struct {
	cmdAble redis.Cmdable
}

func (r *CmdAble) BuildVector(name string, uni, step time.Duration) (resources.Vector, error) {

	if uni < time.Second {
		uni = time.Second
	}
	if step < 500*time.Millisecond {
		step = 500 * time.Millisecond
	}

	size := uni / step
	if size > 20 {
		size = 20
	}
	step = uni / size

	key := fmt.Sprintf("%s:%d:%d", name, uni, step)

	return newVector(key, int64(uni), int64(step), r.cmdAble), nil
}

func (r *CmdAble) Tx() resources.TX {
	tx := r.cmdAble.TxPipeline()
	return &TxPipeline{
		CmdAble: CmdAble{
			cmdAble: tx,
		},
		p: tx,
	}
}

// 加锁：使用 ExponentialBackOff 重试
func acquireLockWithBackoff(ctx context.Context, rdb redis.Cmdable, key, value string, ttl int) (bool, error) {
	operation := func() error {
		res, err := rdb.SetNX(ctx, key, value, time.Duration(ttl)*time.Second).Result()
		if err != nil {
			return err // 网络错误，重试
		}
		if res {
			return nil // 成功，不再重试
		}
		return backoff.Permanent(fmt.Errorf("锁已被占用")) // 加锁失败，但非永久错误（继续重试）
	}

	// 配置指数退避：初始 100ms，乘数 2，最大间隔 5s，最大时长 30s，抖动 0.5
	b := backoff.NewExponentialBackOff()
	b.InitialInterval = 100 * time.Millisecond
	b.Multiplier = 2
	b.MaxInterval = 5 * time.Second
	b.MaxElapsedTime = 30 * time.Second
	b.RandomizationFactor = 0.5 // 抖动：间隔 ±50%

	err := backoff.Retry(operation, backoff.WithContext(b, ctx))
	if err != nil {
		return false, err
	}
	return true, nil
}

func (r *CmdAble) AcquireLock(ctx context.Context, key string, value string, ttl int) resources.BoolResult {
	ok, err := acquireLockWithBackoff(ctx, r.cmdAble, key, value, ttl)
	return newBoolResult(ok, err)
}

// Lua 解锁脚本
var unlockScript = redis.NewScript(`
if redis.call("get", KEYS[1]) == ARGV[1] then
    return redis.call("del", KEYS[1])
else
    return 0
end
`)

func (r *CmdAble) ReleaseLock(ctx context.Context, key string, value string) resources.StatusResult {
	return &statusResult{err: unlockScript.Run(ctx, r.cmdAble, []string{key}, value).Err()}
}

func (r *CmdAble) Set(ctx context.Context, key string, value []byte, expiration time.Duration) resources.StatusResult {

	return &statusResult{err: r.cmdAble.Set(ctx, key, value, expiration).Err()}
}

func (r *CmdAble) SetNX(ctx context.Context, key string, value []byte, expiration time.Duration) resources.BoolResult {

	return r.cmdAble.SetNX(ctx, key, value, expiration)
}

func (r *CmdAble) DecrBy(ctx context.Context, key string, decrement int64, expiration time.Duration) resources.IntResult {
	pipeline := r.cmdAble.Pipeline()
	result := pipeline.DecrBy(ctx, key, decrement)
	if expiration > 0 {
		pipeline.Expire(ctx, key, expiration)
	}
	_, err := pipeline.Exec(ctx)
	if err != nil {
		return nil
	}
	return result

}

func (r *CmdAble) IncrBy(ctx context.Context, key string, decrement int64, expiration time.Duration) resources.IntResult {
	pipeline := r.cmdAble.Pipeline()
	result := pipeline.IncrBy(ctx, key, decrement)
	if expiration > 0 {
		pipeline.Expire(ctx, key, expiration)
	}

	_, err := pipeline.Exec(ctx)
	if err != nil {
		return nil
	}
	return result
}

func (r *CmdAble) Keys(ctx context.Context, key string) resources.StringSliceResult {
	return r.cmdAble.Keys(ctx, key)
}

func (r *CmdAble) Get(ctx context.Context, key string) resources.StringResult {
	return r.cmdAble.Get(ctx, key)

}

func (r *CmdAble) GetDel(ctx context.Context, key string) resources.StringResult {
	return r.cmdAble.GetDel(ctx, key)

}

func (r *CmdAble) HMSetN(ctx context.Context, key string, fields map[string]interface{}, expiration time.Duration) resources.BoolResult {
	pipeline := r.cmdAble.Pipeline()
	result := pipeline.HMSet(ctx, key, fields)
	if expiration > 0 {
		pipeline.Expire(ctx, key, expiration)
	}
	_, err := pipeline.Exec(ctx)
	if err != nil {
		log.Errorf("HMSetN error:%s", err.Error())
		return nil
	}
	return result
}

func (r *CmdAble) HMGet(ctx context.Context, key string, fields ...string) resources.ArrayInterfaceResult {
	return r.cmdAble.HMGet(ctx, key, fields...)
}

func (r *CmdAble) Del(ctx context.Context, keys ...string) resources.IntResult {
	return r.cmdAble.Del(ctx, keys...)
}

func (r *CmdAble) Run(ctx context.Context, script interface{}, keys []string, args ...interface{}) resources.InterfaceResult {
	switch s := script.(type) {
	case string:
		return redis.NewScript(s).Run(ctx, r.cmdAble, keys, args...)
	case *redis.Script:
		return s.Run(ctx, r.cmdAble, keys, args...)
	}
	return resources.NewInterfaceResult(nil, fmt.Errorf("script type error: %T", script))
}

type TxPipeline struct {
	CmdAble
	p redis.Pipeliner
}

func (tx *TxPipeline) Tx() resources.TX {
	return tx
}
func (tx *TxPipeline) Exec(ctx context.Context) error {
	_, err := tx.p.Exec(ctx)

	return err

}

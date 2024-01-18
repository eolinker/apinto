package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/eolinker/eosc/log"

	"github.com/eolinker/apinto/resources"
	"github.com/go-redis/redis/v8"
)

type statusResult struct {
	statusCmd *redis.StatusCmd
}

func (s *statusResult) Result() error {
	return s.statusCmd.Err()
}

type Cmdable struct {
	cmdable redis.Cmdable
}

func (r *Cmdable) BuildVector(name string, uni, step time.Duration) (resources.Vector, error) {

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

	return newVector(key, int64(uni), int64(step), r.cmdable), nil
}

func (r *Cmdable) Tx() resources.TX {
	tx := r.cmdable.TxPipeline()
	return &TxPipeline{
		Cmdable: Cmdable{
			cmdable: tx,
		},
		p: tx,
	}
}

func (r *Cmdable) Set(ctx context.Context, key string, value []byte, expiration time.Duration) resources.StatusResult {

	return &statusResult{statusCmd: r.cmdable.Set(ctx, key, value, expiration)}
}

func (r *Cmdable) SetNX(ctx context.Context, key string, value []byte, expiration time.Duration) resources.BoolResult {

	return r.cmdable.SetNX(ctx, key, value, expiration)
}

func (r *Cmdable) DecrBy(ctx context.Context, key string, decrement int64, expiration time.Duration) resources.IntResult {
	pipeline := r.cmdable.Pipeline()
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

func (r *Cmdable) IncrBy(ctx context.Context, key string, decrement int64, expiration time.Duration) resources.IntResult {
	pipeline := r.cmdable.Pipeline()
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

func (r *Cmdable) Get(ctx context.Context, key string) resources.StringResult {
	return r.cmdable.Get(ctx, key)

}

func (r *Cmdable) GetDel(ctx context.Context, key string) resources.StringResult {
	return r.cmdable.GetDel(ctx, key)

}

func (r *Cmdable) HMSetN(ctx context.Context, key string, fields map[string]interface{}, expiration time.Duration) resources.BoolResult {
	pipeline := r.cmdable.Pipeline()
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

func (r *Cmdable) HMGet(ctx context.Context, key string, fields ...string) resources.ArrayInterfaceResult {
	return r.cmdable.HMGet(ctx, key, fields...)
}

func (r *Cmdable) Del(ctx context.Context, keys ...string) resources.IntResult {
	return r.cmdable.Del(ctx, keys...)
}

func (r *Cmdable) Run(ctx context.Context, script interface{}, keys []string, args ...interface{}) resources.InterfaceResult {
	switch s := script.(type) {
	case string:
		return redis.NewScript(s).Run(ctx, r.cmdable, keys, args...)
	case *redis.Script:
		return s.Run(ctx, r.cmdable, keys, args...)
	}
	return resources.NewInterfaceResult(nil, fmt.Errorf("script type error: %T", script))
}

type TxPipeline struct {
	Cmdable
	p redis.Pipeliner
}

func (tx *TxPipeline) Tx() resources.TX {
	return tx
}
func (tx *TxPipeline) Exec(ctx context.Context) error {
	_, err := tx.p.Exec(ctx)

	return err

}

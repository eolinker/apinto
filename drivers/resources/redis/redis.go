package redis

import (
	"context"
	"github.com/eolinker/apinto/resources"
	"github.com/go-redis/redis/v8"
	"time"
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

func (r *Cmdable) DecrBy(ctx context.Context, key string, decrement int64) resources.IntResult {

	return r.cmdable.DecrBy(ctx, key, decrement)
}

func (r *Cmdable) IncrBy(ctx context.Context, key string, decrement int64) resources.IntResult {
	return r.cmdable.IncrBy(ctx, key, decrement)
}

func (r *Cmdable) Get(ctx context.Context, key string) resources.StringResult {
	return r.cmdable.Get(ctx, key)

}

func (r *Cmdable) GetDel(ctx context.Context, key string) resources.StringResult {
	return r.cmdable.GetDel(ctx, key)

}

func (r *Cmdable) Del(ctx context.Context, keys ...string) resources.IntResult {
	return r.cmdable.Del(ctx, keys...)
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

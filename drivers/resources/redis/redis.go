package redis

import (
	"context"
	"github.com/go-redis/redis/v8"
	"time"
)

type _Cacher struct {
	client *redis.ClusterClient
}

func (r *_Cacher) Set(ctx context.Context, key string, value []byte, expiration time.Duration) error {

	return r.client.Set(ctx, key, value, expiration).Err()
}

func (r *_Cacher) SetNX(ctx context.Context, key string, value []byte, expiration time.Duration) (bool, error) {

	return r.client.SetNX(ctx, key, value, expiration).Result()
}

func (r *_Cacher) DecrBy(ctx context.Context, key string, decrement int64) (int64, error) {

	return r.client.DecrBy(ctx, key, decrement).Result()
}

func (r *_Cacher) IncrBy(ctx context.Context, key string, decrement int64) (int64, error) {
	return r.client.IncrBy(ctx, key, decrement).Result()
}

func (r *_Cacher) Get(ctx context.Context, key string) ([]byte, error) {
	return r.client.Get(ctx, key).Bytes()

}

func (r *_Cacher) GetDel(ctx context.Context, key string) ([]byte, error) {
	return r.client.GetDel(ctx, key).Bytes()

}

func (r *_Cacher) Del(ctx context.Context, keys ...string) (int64, error) {
	return r.client.Del(ctx, keys...).Result()
}

func newCacher(client *redis.ClusterClient) *_Cacher {
	if client == nil {
		return nil
	}
	return &_Cacher{client: client}
}

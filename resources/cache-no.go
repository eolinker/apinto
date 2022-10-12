package resources

import (
	"context"
	"time"
)

type NoCache struct {
}

func (n *NoCache) Set(ctx context.Context, key string, value []byte, expiration time.Duration) error {
	return ErrorNoCache
}

func (n *NoCache) SetNX(ctx context.Context, key string, value []byte, expiration time.Duration) (bool, error) {
	return false, ErrorNoCache
}

func (n *NoCache) DecrBy(ctx context.Context, key string, decrement int64) (int64, error) {
	return 0, ErrorNoCache
}

func (n *NoCache) IncrBy(ctx context.Context, key string, decrement int64) (int64, error) {
	return 0, ErrorNoCache

}

func (n *NoCache) Get(ctx context.Context, key string) ([]byte, error) {
	return nil, ErrorNoCache

}

func (n *NoCache) GetDel(ctx context.Context, key string) ([]byte, error) {
	return nil, ErrorNoCache
}

func (n *NoCache) Del(ctx context.Context, keys ...string) (int64, error) {
	return 0, ErrorNoCache
}

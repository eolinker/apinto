package resources

import (
	"context"
	"github.com/coocood/freecache"
	"time"
)

type NoCache struct {
	client *freecache.Cache
}

func (n *NoCache) Set(ctx context.Context, key string, value []byte, expiration time.Duration) error {

	return n.client.Set([]byte(key), value, int(expiration.Seconds()))
}

func (n *NoCache) SetNX(ctx context.Context, key string, value []byte, expiration time.Duration) (bool, error) {

	_, err := n.client.GetOrSet([]byte(key), value, int(expiration.Seconds()))
	if err != nil {
		return false, err
	}

	return true, nil
}

func (n *NoCache) DecrBy(ctx context.Context, key string, decrement int64) (int64, error) {
	return 0, ErrorNoCache
}

func (n *NoCache) IncrBy(ctx context.Context, key string, decrement int64) (int64, error) {
	return 0, ErrorNoCache

}

func (n *NoCache) Get(ctx context.Context, key string) ([]byte, error) {
	return n.client.Get([]byte(key))

}

func (n *NoCache) GetDel(ctx context.Context, key string) ([]byte, error) {
	bytes, err := n.client.Get([]byte(key))
	if err != nil {
		return nil, err
	}
	n.client.Del([]byte(key))
	return bytes, nil
}

func (n *NoCache) Del(ctx context.Context, keys ...string) (int64, error) {

	for _, key := range keys {
		n.client.Del([]byte(key))
	}

	return 1, nil
}

func NewCacher() *NoCache {
	return &NoCache{client: freecache.NewCache(0)}
}

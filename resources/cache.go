package resources

import (
	"context"
	"errors"
	"time"
)

var (
	ErrorNoCache        = errors.New("no cache")
	_            ICache = (*_Proxy)(nil)
)
var (
	singCacheProxy *_Proxy
)

func init() {
	singCacheProxy = newProxy(new(NoCache))
}
func ReplaceCacher(caches ...ICache) {
	if len(caches) < 1 || caches[0] == nil {
		if singCacheProxy.ICache != nil {
			singCacheProxy.ICache.Close()
		}
		singCacheProxy.ICache = NewCacher()
		return
	}
	singCacheProxy.ICache = caches[0]
}

func Cacher() ICache {
	return singCacheProxy
}

type ICache interface {
	Set(ctx context.Context, key string, value []byte, expiration time.Duration) error
	SetNX(ctx context.Context, key string, value []byte, expiration time.Duration) (bool, error)
	DecrBy(ctx context.Context, key string, decrement int64) (int64, error)
	IncrBy(ctx context.Context, key string, decrement int64) (int64, error)
	Get(ctx context.Context, key string) ([]byte, error)
	GetDel(ctx context.Context, key string) ([]byte, error)
	Del(ctx context.Context, keys ...string) (int64, error)
	Close() error
}

type _Proxy struct {
	ICache
}

func newProxy(target ICache) *_Proxy {
	return &_Proxy{ICache: target}
}

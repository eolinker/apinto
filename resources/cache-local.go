package resources

import (
	"context"
	"errors"
	"strconv"
	"sync"
	"time"

	"github.com/coocood/freecache"
)

var (
	localCache ICache = (*cacheLocal)(nil)
)
var (
	once       sync.Once
	LocalCache func() ICache
)

func init() {
	LocalCache = func() ICache {

		once.Do(func() {
			localCache = newCacher()
			LocalCache = func() ICache {
				return localCache
			}
		})
		return localCache
	}

}

type cacheLocal struct {
	txLock sync.Mutex

	keyLock  sync.Mutex
	keyLocks map[string]*sync.Mutex
	client   *freecache.Cache
}
type cacheLocalTX struct {
	*cacheLocal
}

func (n *cacheLocalTX) Tx() TX {
	return n
}
func (n *cacheLocal) Tx() TX {
	n.txLock.Lock()
	return &cacheLocalTX{cacheLocal: n}
}

func (n *cacheLocal) Exec(ctx context.Context) error {
	n.txLock.Unlock()
	return nil
}

func (n *cacheLocal) Close() error {
	n.client.Clear()
	return nil
}

func (n *cacheLocal) AcquireLock(ctx context.Context, key string, value string, ttl int) BoolResult {

	n.keyLock.Lock()
	lock, has := n.keyLocks[key]
	if !has {
		lock = &sync.Mutex{}
		n.keyLocks[key] = lock
	}
	n.keyLock.Unlock()
	lock.Lock()
	return NewBoolResult(true, nil)
}

func (n *cacheLocal) ReleaseLock(ctx context.Context, key string, value string) StatusResult {
	n.keyLock.Lock()
	lock, has := n.keyLocks[key]
	if has {
		lock.Unlock()
		delete(n.keyLocks, key)
	}
	n.keyLock.Unlock()
	return NewStatusResult(nil)
}

func (n *cacheLocal) Set(ctx context.Context, key string, value []byte, expiration time.Duration) StatusResult {

	err := n.client.Set([]byte(key), value, int(expiration.Seconds()))
	return NewStatusResult(err)
}

func (n *cacheLocal) SetNX(ctx context.Context, key string, value []byte, expiration time.Duration) BoolResult {

	old, err := n.client.GetOrSet([]byte(key), value, int(expiration.Seconds()))
	if err != nil {
		return NewBoolResult(false, err)
	}
	return NewBoolResult(old == nil, nil)
}

func (n *cacheLocal) DecrBy(ctx context.Context, key string, decrement int64, expiration time.Duration) IntResult {
	return n.IncrBy(ctx, key, -decrement, expiration)
}

func (n *cacheLocal) IncrBy(ctx context.Context, key string, incr int64, expiration time.Duration) IntResult {

	n.keyLock.Lock()
	lock, has := n.keyLocks[key]
	if !has {
		lock = new(sync.Mutex)
		n.keyLocks[key] = lock
		lock.Lock()
	}
	n.keyLock.Unlock()

	if has {
		lock.Lock()
	}
	defer func() {
		lock.Unlock()
		if n.keyLock.TryLock() {
			if lock.TryLock() {
				delete(n.keyLocks, key)
				lock.Unlock()
			}
			n.keyLock.Unlock()
		}
	}()

	v, err := n.client.Get([]byte(key))
	if err != nil {
		v = ToBytes(incr)
		err := n.client.Set([]byte(key), v, int(expiration.Seconds()))
		if err != nil {
			return NewIntResult(0, err)
		}
		return NewIntResult(incr, nil)
	}
	value := ToInt(v) + incr
	v = ToBytes(value)
	err = n.client.Set([]byte(key), v, int(expiration.Seconds()))
	if err != nil {
		return NewIntResult(0, err)
	}

	return NewIntResult(value, nil)
}
func ToInt(b []byte) int64 {
	v, err := strconv.ParseInt(string(b), 10, 64)
	if err != nil {
		return 0
	}
	return v
}
func ToBytes(v int64) []byte {
	return []byte(strconv.FormatInt(v, 10))
}

func (n *cacheLocal) Keys(ctx context.Context, pattern string) StringSliceResult {
	return NewStringSliceResult(nil, errors.New("not support"))
}

func (n *cacheLocal) Get(ctx context.Context, key string) StringResult {
	data, err := n.client.Get([]byte(key))
	if err != nil {
		return NewStringResult("", err)
	}
	return NewStringResultBytes(data, err)

}

func (n *cacheLocal) GetDel(ctx context.Context, key string) StringResult {
	bytes, err := n.client.Get([]byte(key))
	if err != nil {
		return NewStringResult("", err)
	}
	n.client.Del([]byte(key))
	return NewStringResultBytes(bytes, nil)
}

func (n *cacheLocal) HMSetN(ctx context.Context, key string, fields map[string]interface{}, expiration time.Duration) BoolResult {
	return NewBoolResult(false, errors.New("not support"))
}

func (n *cacheLocal) HMGet(ctx context.Context, key string, fields ...string) ArrayInterfaceResult {
	return NewArrayInterfaceResult(nil, errors.New("not support"))
}

func (n *cacheLocal) Del(ctx context.Context, keys ...string) IntResult {
	var count int64 = 0
	for _, key := range keys {
		if n.client.Del([]byte(key)) {
			count++
		}
	}

	return NewIntResult(count, nil)
}

func (n *cacheLocal) Run(ctx context.Context, script interface{}, keys []string, args ...interface{}) InterfaceResult {
	return NewInterfaceResult(nil, errors.New("not support"))
}

func newCacher() *cacheLocal {
	return &cacheLocal{client: freecache.NewCache(2048 * 1024), keyLocks: make(map[string]*sync.Mutex)}
}

package resources

import (
	"context"
	"encoding/binary"
	"github.com/coocood/freecache"
	"sync"
	"time"
)

var (
	_ ICache = (*NoCache)(nil)
)

type NoCache struct {
	txLock sync.Mutex

	keyLock  sync.Mutex
	keyLocks map[string]*sync.Mutex
	client   *freecache.Cache
}
type NoCacheTX struct {
	*NoCache
}

func (n *NoCacheTX) Tx() TX {
	return n
}
func (n *NoCache) Tx() TX {
	n.txLock.Lock()
	return &NoCacheTX{NoCache: n}
}

func (n *NoCache) Exec(ctx context.Context) error {
	n.txLock.Unlock()
	return nil
}

func (n *NoCache) Close() error {
	n.client.Clear()
	return nil
}

func (n *NoCache) Set(ctx context.Context, key string, value []byte, expiration time.Duration) StatusResult {

	err := n.client.Set([]byte(key), value, int(expiration.Seconds()))
	return NewStatusResult(err)
}

func (n *NoCache) SetNX(ctx context.Context, key string, value []byte, expiration time.Duration) BoolResult {

	old, err := n.client.GetOrSet([]byte(key), value, int(expiration.Seconds()))
	if err != nil {
		return NewBoolResult(false, err)
	}
	return NewBoolResult(old == nil, nil)
}

func (n *NoCache) DecrBy(ctx context.Context, key string, decrement int64, expiration time.Duration) IntResult {
	return n.IncrBy(ctx, key, -decrement, expiration)
}

func (n *NoCache) IncrBy(ctx context.Context, key string, incr int64, expiration time.Duration) IntResult {
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
	if err != nil || len(v) != 8 {
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
	return int64(binary.LittleEndian.Uint64(b))
}
func ToBytes(v int64) []byte {
	var b [8]byte
	binary.LittleEndian.PutUint64(b[:], uint64(v))
	return b[:]
}
func (n *NoCache) Get(ctx context.Context, key string) StringResult {
	data, err := n.client.Get([]byte(key))
	if err != nil {
		return nil
	}
	return NewStringResultBytes(data, err)

}

func (n *NoCache) GetDel(ctx context.Context, key string) StringResult {
	bytes, err := n.client.Get([]byte(key))
	if err != nil {
		return NewStringResult("", err)
	}
	n.client.Del([]byte(key))
	return NewStringResultBytes(bytes, nil)
}

func (n *NoCache) Del(ctx context.Context, keys ...string) IntResult {
	var count int64 = 0
	for _, key := range keys {
		if n.client.Del([]byte(key)) {
			count++
		}
	}

	return NewIntResult(count, nil)
}

func NewCacher() *NoCache {
	return &NoCache{client: freecache.NewCache(0), keyLocks: make(map[string]*sync.Mutex)}
}

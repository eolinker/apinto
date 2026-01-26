package redis

import (
	"context"
	"fmt"
	"github.com/eolinker/eosc/log"

	redis "github.com/redis/go-redis/v9"
	"strconv"
	"time"
)

const defaultLockTTL = time.Second * 3

type Vector struct {
	name string

	step int64 // 时间步长（纳秒）
	size int64 // 窗口桶大小（槽数）
	cmd  redis.Cmdable
	// 新增：锁 TTL（秒），建议短于业务时间
	lockTTL time.Duration
}

func (v *Vector) CompareAndAdd(key string, threshold, delta int64) (int64, bool) {
	token := fmt.Sprintf("strategy-limiting:%s:%s", v.name, key)
	index := time.Now().UnixNano() / v.step
	bucketStart := (index / v.size) * v.size // 当前桶起始 index
	ctx := context.Background()

	lockKey := fmt.Sprintf("lock:%s", token) // 锁 key 基于 token
	lock := newSimpleLock(v.cmd, lockKey, v.lockTTL)

	if err := lock.Lock(ctx); err != nil {
		// 加锁失败，直接拒绝（可加重试逻辑）
		return 0, false
	}
	defer func() {
		if err := lock.Unlock(ctx); err != nil {
			log.Errorf("Unlock error: %v", err)
		}
		// 可选：清理过期桶（e.g., DEL token if old）
	}()

	// 锁内：安全执行 get + incr
	currentSum := v.get(ctx, token, bucketStart)
	if currentSum > threshold {
		return currentSum, false // 已超阈值
	}

	// 增量（HIncrBy 是原子的，但因锁保护，整个操作安全）
	field := fmt.Sprintf("%d", index)
	_, err := v.cmd.HIncrBy(ctx, token, field, delta).Result()
	if err != nil {
		log.Errorf("HIncrBy error: %v", err)
		return currentSum, false
	}

	return currentSum + delta, true
}

func (v *Vector) Add(key string, delta int64) int64 {
	token := fmt.Sprint("strategy-limiting:", v.name, ":", key)
	index := time.Now().UnixNano() / v.step
	ctx := context.Background()
	result, err := v.cmd.HIncrBy(ctx, token, fmt.Sprint(index), delta).Result()
	if err != nil {
		log.Errorf("redis vector add error %v", err)
	}
	return result
}

func (v *Vector) Get(key string) int64 {
	token := fmt.Sprint("strategy-limiting:", v.name, ":", key)
	index := time.Now().UnixNano() / v.step
	ctx := context.Background()
	from := index / v.size * v.size
	return v.get(ctx, token, from)
}
func (v *Vector) get(ctx context.Context, token string, from int64) int64 {
	result, err := v.cmd.HGetAll(ctx, token).Result()
	if err != nil {
		return 0
	}
	rv := int64(0)
	delKeys := make([]string, 0, len(result))
	for k, v := range result {
		i, e := strconv.ParseInt(k, 10, 64)
		if e != nil || i < from {
			delKeys = append(delKeys, k)
			continue
		}
		value, er := strconv.ParseInt(v, 10, 64)
		if er != nil {
			delKeys = append(delKeys, k)
			continue
		}
		rv += value
	}
	v.cmd.HDel(ctx, token, delKeys...)
	return rv
}

func newVector(name string, uin int64, step int64, cmd redis.Cmdable) *Vector {
	return &Vector{name: name, step: step, cmd: cmd, size: uin / step, lockTTL: defaultLockTTL}
}

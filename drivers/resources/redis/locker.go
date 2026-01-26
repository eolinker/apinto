package redis

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	redis "github.com/redis/go-redis/v9"
	"time"
)

// 锁结构体（复用第一种方式的简单实现）
type simpleLock struct {
	client redis.Cmdable // 假设 cmd 是 Client；若 Cmdable，需适配
	key    string
	value  string
	ttl    time.Duration
}

func newSimpleLock(client redis.Cmdable, key string, ttl time.Duration) *simpleLock {
	return &simpleLock{
		client: client,
		key:    key,
		ttl:    ttl,
	}
}

func (sl *simpleLock) Lock(ctx context.Context) error {
	sl.value = uuid.New().String()
	setCmd := sl.client.Set(ctx, sl.key, sl.value, sl.ttl)
	return setCmd.Err() // 若键存在，返回 Err
}

func (sl *simpleLock) Unlock(ctx context.Context) error {
	luaScript := `
	if redis.call("GET", KEYS[1]) == ARGV[1] then
		return redis.call("DEL", KEYS[1])
	else
		return 0
	end
	`
	res, err := sl.client.Eval(ctx, luaScript, []string{sl.key}, sl.value).Result()
	if err != nil {
		return err
	}
	if res.(int64) == 0 {
		return fmt.Errorf("lock not owned")
	}
	return nil
}

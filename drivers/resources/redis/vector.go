package redis

import (
	"context"
	"fmt"
	"github.com/go-redis/redis/v8"
	"strconv"
	"time"
)

type Vector struct {
	name string

	step         int64
	size         int64
	redisCmdable redis.Cmdable
}

func (v *Vector) Add(key string, delta int64) int64 {
	token := fmt.Sprint(v.name, ":", key)
	index := time.Now().UnixNano() / v.step
	ctx := context.Background()
	v.redisCmdable.HIncrBy(ctx, token, fmt.Sprint(index), delta)
	return v.get(ctx, token, index/v.size*v.size)
}

func (v *Vector) Get(key string) int64 {
	token := fmt.Sprint(v.name, ":", key)
	index := time.Now().UnixNano() / v.step
	ctx := context.Background()
	return v.get(ctx, token, index/v.size*v.size)
}
func (v *Vector) get(ctx context.Context, token string, from int64) int64 {
	result, err := v.redisCmdable.HGetAll(ctx, token).Result()
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
	v.redisCmdable.HDel(ctx, token, delKeys...)
	return rv
}

func newVector(name string, uin int64, step int64, redisCmdable redis.Cmdable) *Vector {
	return &Vector{name: name, step: step, redisCmdable: redisCmdable, size: uin / step}
}

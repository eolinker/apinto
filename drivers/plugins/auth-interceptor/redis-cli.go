package auth_interceptor

import (
	"context"
	"strings"
	"sync/atomic"
	"time"

	"github.com/eolinker/eosc"

	redis "github.com/redis/go-redis/v9"
)

var redisPool = &pool{
	pool: eosc.BuildUntyped[string, *redisClient](),
}

type pool struct {
	pool eosc.Untyped[string, *redisClient]
}

func (p *pool) Use(name string) bool {
	_, ok := p.pool.Get(name)
	if !ok {
		return false
	}
	return true
}

func (p *pool) Get(name string) redis.UniversalClient {
	client, ok := p.pool.Get(name)
	if !ok {
		return nil
	}
	return client.Get()
}

func (p *pool) Set(name string, client redis.UniversalClient) {
	c, ok := p.pool.Get(name)
	if !ok {
		c = &redisClient{
			use: 1,
		}
	}
	c.client = client
	p.pool.Set(name, c)
}

func (p *pool) Release(name string) {
	client, ok := p.pool.Get(name)
	if !ok {
		return
	}
	if client.Release() == 0 {
		client.Close()
		p.pool.Del(name)
	}
}

type redisClient struct {
	client redis.UniversalClient
	use    int32
}

func (r *redisClient) Use() {
	atomic.AddInt32(&r.use, 1)
	return
}

func (r *redisClient) Get() redis.UniversalClient {
	return r.client
}

func (r *redisClient) Set(client redis.UniversalClient) {
	r.client = client
}

func (r *redisClient) Release() int32 {
	n := atomic.AddInt32(&r.use, -1)
	return n
}

func (r *redisClient) Close() {
	if r.client == nil {
		return
	}
	r.client.Close()
	r.client = nil
}

func initRedis(cfg *RedisConfig) (redis.UniversalClient, error) {
	var client redis.UniversalClient
	switch cfg.Mode {
	case "cluster":
		client = redis.NewClusterClient(&redis.ClusterOptions{
			Addrs:    strings.Split(cfg.Addr, ","),
			Username: cfg.Username,
			Password: cfg.Password,
		})
	default:
		client = redis.NewClient(&redis.Options{
			Addr:     cfg.Addr,
			Username: cfg.Username,
			Password: cfg.Password,
			DB:       cfg.DB,
		})
	}

	timeout, cancelFunc := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelFunc()
	if err := client.Ping(timeout).Err(); err != nil {
		_ = client.Close()
		return nil, err
	}
	return client, nil
}

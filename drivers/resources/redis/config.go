package redis

import (
	"context"
	"fmt"
	"github.com/eolinker/eosc"
	"github.com/go-redis/redis/v8"
	"time"
)

type Config struct {
	Addrs    []string `json:"addrs" label:"redis 节点列表"`
	Username string   `json:"username"`
	Password string   `json:"password"`
	Scopes   []string `json:"scopes" label:"资源组"`
}

func (c *Config) connect() (*redis.ClusterClient, error) {
	if len(c.Addrs) == 0 {
		return nil, fmt.Errorf("addrs:%w", eosc.ErrorRequire)
	}
	nc := redis.NewClusterClient(&redis.ClusterOptions{
		Addrs:    c.Addrs,
		Username: c.Username,
		Password: c.Password,
	})
	timeout, _ := context.WithTimeout(context.Background(), time.Second)
	if err := nc.Ping(timeout).Err(); err != nil {
		nc.Close()
		return nil, err
	}
	return nc, nil
}

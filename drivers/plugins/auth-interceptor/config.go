package auth_interceptor

import (
	"fmt"
	"strings"

	"github.com/eolinker/apinto/drivers"
	"github.com/eolinker/eosc"
	redis "github.com/redis/go-redis/v9"
)

const (
	positionQuery  = "query"
	positionHeader = "header"
	positionBody   = "body"
)

// Config 参数配置属性
type Config struct {
	SysKey       string       `json:"sys_key" label:"认证系统在Redis存储中的KEY值"`
	AuthKey      string       `json:"auth_key" label:"认证参数KEY名"`
	AuthPosition string       `json:"auth_position" enum:"header,query,body" label:"认证参数追加位置"`
	RedisConfig  *RedisConfig `json:"redis" label:"Redis连接配置"`
	RedisConn    string       `json:"redis_conn" label:"Redis连接名称"`
	RetryCount   int          `json:"retry_count" label:"认证失败重试次数"`
	RetryPeriod  int          `json:"retry_period" label:"认证失败重试间隔"`
}

type RedisConfig struct {
	Addr     string `json:"addr" label:"Redis连接地址"`
	Username string `json:"username" label:"Redis连接用户名"`
	Password string `json:"password" label:"Redis连接密码"`
	DB       int    `json:"db" label:"Redis db序号"`
	Mode     string `json:"mode" label:"Redis连接模式"`
}

// Create 初始化插件执行实例
func Create(id, name string, conf *Config, workers map[eosc.RequireId]eosc.IWorker) (eosc.IWorker, error) {
	err := conf.doCheck()
	if err != nil {
		return nil, err
	}
	var client redis.UniversalClient
	ok := redisPool.Use(conf.RedisConn)
	if !ok {
		client, err = initRedis(conf.RedisConfig)
		if err != nil {
			return nil, err
		}
		redisPool.Set(conf.RedisConn, client)
	}
	auth := Auth{
		WorkerBase: drivers.Worker(id, name),
		cfg:        conf,
		redisConn:  conf.RedisConn,
	}

	return &auth, nil
}

func (c *Config) doCheck() error {
	if c.SysKey == "" {
		return fmt.Errorf("[plugin auth-interceptor config err] param sys_key must be not null")
	}
	c.AuthPosition = strings.ToLower(c.AuthPosition)

	switch c.AuthPosition {
	case positionQuery, positionHeader, positionBody:
	default:
		return fmt.Errorf(`[plugin auth-interceptor config err] param position must be in the set ["query","header",body]. err position: %s `, c.AuthPosition)

	}

	if c.AuthKey == "" {
		return fmt.Errorf("[plugin auth-interceptor config err] param auth_key must be not null")
	}

	if c.RedisConn == "" {
		return fmt.Errorf("[plugin auth-interceptor config err] param redis_conn must be not null")
	}
	if c.RedisConfig == nil || c.RedisConfig.Addr == "" {
		return fmt.Errorf("[plugin auth-interceptor config err] param RedisAddress must be not null")
	}

	if c.RetryCount < 0 {
		c.RetryCount = 0
	}

	if c.RetryPeriod < 1 {
		c.RetryPeriod = 5
	}

	return nil
}

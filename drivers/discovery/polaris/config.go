package polaris

import (
	"github.com/polarismesh/polaris-go"
)

// Config 北极星驱动配置
type Config struct {
	Config AccessConfig `json:"config" label:"配置信息"`
}

// AccessConfig 接入地址配置
type AccessConfig struct {
	Address   []string          `json:"address" label:"北极星地址"`
	Namespace string            `json:"namespace" label:"命名空间"`
	Params    map[string]string `json:"params" label:"参数"`
}

// polarisClient 北极星主调端
type polarisClients struct {
	consumerAPI polaris.ConsumerAPI
	namespace   string
}

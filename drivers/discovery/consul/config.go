package consul

import (
	"github.com/hashicorp/consul/api"
)

const defaultScheme = "http"

// Config consul驱动配置
type Config struct {
	Config AccessConfig `json:"config" label:"配置信息"`
}

// AccessConfig 接入地址配置
type AccessConfig struct {
	Address []string          `json:"address" label:"consul地址"`
	Params  map[string]string `json:"params" label:"参数"`
}

type consulClients struct {
	clients []*api.Client
}

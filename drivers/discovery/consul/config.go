package consul

import (
	"strings"

	"github.com/hashicorp/consul/api"
)

//Config consul驱动配置
type Config struct {
	Scheme string       `json:"scheme" label:"请求协议" enum:"HTTP,HTTPS"`
	Config AccessConfig `json:"config" label:"配置信息"`
}

//AccessConfig 接入地址配置
type AccessConfig struct {
	Address []string          `json:"address" label:"consul地址"`
	Params  map[string]string `json:"params" label:"参数"`
}

type consulClients struct {
	clients []*api.Client
}

func (c *Config) getScheme() string {
	scheme := strings.ToLower(c.Scheme)
	if scheme != "http" && scheme != "https" {
		scheme = "http"
	}
	return scheme
}

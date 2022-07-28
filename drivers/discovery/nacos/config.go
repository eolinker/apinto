package nacos

import (
	"net/url"
)

const defaultScheme = "http"

//Config nacos驱动配置
type Config struct {
	Scheme string       `json:"scheme" label:"请求协议" enum:"HTTP,HTTPS" skip:""`
	Config AccessConfig `json:"config" label:"配置信息"`
}

//AccessConfig 接入地址配置
type AccessConfig struct {
	Address []string          `json:"address" label:"nacos地址"`
	Params  map[string]string `json:"params" label:"参数"`
}

func (c *Config) getParams() url.Values {
	p := url.Values{}
	p.Set("healthyOnly", "true")
	for k, v := range c.Config.Params {
		p.Set(k, v)
	}
	return p
}

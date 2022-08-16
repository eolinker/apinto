package service

import (
	"encoding/json"
	"strings"

	"github.com/eolinker/apinto/plugin"

	"github.com/eolinker/eosc"
)

//Config service_http驱动配置
type Config struct {
	Timeout   int64          `json:"timeout" label:"请求超时时间" default:"2000" minimum:"1" description:"单位：ms，最小值：1"`
	Retry     int            `json:"retry" label:"失败重试次数"`
	Scheme    string         `json:"scheme" label:"请求协议" enum:"HTTP,HTTPS"`
	Discovery eosc.RequireId `json:"discovery" required:"false" empty_label:"使用匿名上游" label:"服务发现" skill:"github.com/eolinker/apinto/discovery.discovery.IDiscovery"`
	Service   string         `json:"service" required:"false" label:"服务名 or 配置" switch:"discovery !==''"`
	Nodes     []string       `json:"nodes" label:"静态配置" switch:"discovery===''"`
	Balance   string         `json:"balance" enum:"round-robin" label:"负载均衡算法"`
	Plugins   plugin.Plugins `json:"plugins" label:"插件"`
}

func (c *Config) String() string {
	data, _ := json.Marshal(c)
	return string(data)
}
func (c *Config) rebuild() {
	if c.Retry < 0 {
		c.Retry = 0
	}
	if c.Timeout < 0 {
		c.Timeout = 0
	}
	c.Scheme = strings.ToLower(c.Scheme)
	if c.Scheme != "http" && c.Scheme != "https" {
		c.Scheme = "http"
	}

}

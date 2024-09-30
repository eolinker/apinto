package ai_service

import (
	"encoding/json"
	"strings"

	"github.com/eolinker/eosc"
)

// Config service_http驱动配置
type Config struct {
	Title    string         `json:"title" label:"标题"`
	Timeout  int64          `json:"timeout" label:"请求超时时间" default:"2000" minimum:"1" title:"单位：ms，最小值：1"`
	Retry    int            `json:"retry" label:"失败重试次数"`
	Scheme   string         `json:"scheme" label:"请求协议" enum:"HTTP,HTTPS"`
	Provider eosc.RequireId `json:"provider" required:"false" empty_label:"使用匿名上游" label:"服务发现" skill:"github.com/eolinker/apinto/discovery.discovery.IDiscovery"`
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

package service_http

import (
	"strings"

	"github.com/eolinker/apinto/plugin"

	"github.com/eolinker/eosc"
)

type AnonymousConfig struct {
	Type   string `json:"type" enum:"round-robin" label:"负载算法"`
	Config string `json:"config" label:"配置"`
}

//Config service_http驱动配置
type Config struct {
	Timeout           int64                     `json:"timeout" label:"请求超时时间（单位ms）"`
	Retry             int                       `json:"retry" label:"失败重试次数"`
	Scheme            string                    `json:"scheme" label:"请求协议" enum:"HTTP,HTTPS"`
	Upstream          eosc.RequireId            `json:"upstream"  label:"上游" skill:"github.com/eolinker/apinto/upstream.upstream.IUpstream" required:"false" empty_label:"使用匿名上游"`
	UpstreamAnonymous *AnonymousConfig          `json:"anonymous" label:"匿名上游" switch:"upstream===''" `
	PluginConfig      map[string]*plugin.Config `json:"plugins" label:"插件"`
}

var validMethods = []string{
	"GET",
	"POST",
	"PUT",
	"DELETE",
	"PATCH",
	"HEAD",
	"OPTIONS",
}

var validScheme = []string{
	"HTTP",
	"HTTPS",
}

func (c *Config) rebuild() {
	if c.Retry < 0 {
		c.Retry = 0
	}
	if c.Timeout < 0 {
		c.Timeout = 0
	}

	if !checkValidParams(strings.ToUpper(c.Scheme), validScheme) {
		c.Scheme = "http"
	}
}

func checkValidParams(data string, params []string) bool {
	for _, p := range params {
		if data == p {
			return true
		}
	}
	return false
}

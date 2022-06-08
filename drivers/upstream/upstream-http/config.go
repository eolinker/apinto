package upstream_http

import (
	"github.com/eolinker/apinto/plugin"
	"github.com/eolinker/eosc"
)

//Config http-service-proxy驱动配置结构体
type Config struct {
	Desc      string                    `json:"desc" label:"描述"`
	Scheme    string                    `json:"scheme" enum:"HTTP,HTTPS" label:"请求协议"`
	Type      string                    `json:"type" enum:"round-robin" label:"负载算法"`
	Config    string                    `json:"config" label:"配置"`
	Discovery eosc.RequireId            `json:"discovery" label:"服务发现" skill:"github.com/eolinker/apinto/discovery.discovery.IDiscovery"`
	Plugins   map[string]*plugin.Config `json:"plugins" label:"插件"`
}

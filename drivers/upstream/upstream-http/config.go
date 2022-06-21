package upstream_http

import "github.com/eolinker/eosc"

//Config http-service-proxy驱动配置结构体
type Config struct {
	Discovery eosc.RequireId `json:"discovery" required:"true" label:"服务发现" skill:"github.com/eolinker/apinto/discovery.discovery.IDiscovery"`
	Config    string         `json:"config" label:"配置"`
	Scheme    string         `json:"scheme" enum:"HTTP,HTTPS" label:"请求协议"`
	Type      string         `json:"type" enum:"round-robin" label:"负载算法"`
}

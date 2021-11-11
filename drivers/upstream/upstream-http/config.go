package upstream_http

import (
	"github.com/eolinker/eosc"
)

//Config http-service-proxy驱动配置结构体
type Config struct {
	Desc      string         `json:"desc"`
	Scheme    string         `json:"scheme"`
	Type      string         `json:"type"`
	Config    string         `json:"config"`
	Discovery eosc.RequireId `json:"discovery" skill:"github.com/eolinker/goku/discovery.discovery.IDiscovery"`
}

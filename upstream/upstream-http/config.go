package upstream_http

import (
	"github.com/eolinker/eosc"
)

//Config http-proxy驱动配置结构体
type Config struct {
	id        string
	Name      string         `json:"name"`
	Driver    string         `json:"driver"`
	Desc      string         `json:"desc"`
	Scheme    string         `json:"scheme"`
	Type      string         `json:"type"`
	Config    string         `json:"config"`
	Discovery eosc.RequireId `json:"upstream" skill:"github.com/eolinker/goku-eosc/discovery.discovery.IDiscovery"`
}

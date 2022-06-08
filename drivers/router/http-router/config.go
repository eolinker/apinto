package http_router

import (
	"github.com/eolinker/apinto/plugin"
	router_http "github.com/eolinker/apinto/router/router-http"
	"github.com/eolinker/apinto/service"
	"github.com/eolinker/eosc"
)

//DriverConfig http路由驱动配置
type DriverConfig struct {
	Listen  int                       `json:"listen" yaml:"listen" title:"port" description:"使用端口" default:"80" label:"端口号" maximum:"65535"`
	Method  []string                  `json:"method" yaml:"method" enum:"GET,POST,PUT,DELETE,PATH,HEAD,OPTIONS" label:"请求方式"`
	Host    []string                  `json:"host" yaml:"host" label:"域名"`
	Rules   []DriverRule              `json:"rules" yaml:"rules" label:"路由规则"`
	Target  eosc.RequireId            `json:"target" yaml:"target" skill:"github.com/eolinker/apinto/service.service.IService" label:"目标服务"`
	Disable bool                      `json:"disable" yaml:"disable" label:"是否启用"`
	Plugins map[string]*plugin.Config `json:"plugins" yaml:"plugins" label:"插件配置"`
}

//DriverRule http路由驱动配置Rule结构体
type DriverRule struct {
	Location string            `json:"location" yaml:"location"`
	Header   map[string]string `json:"header" yaml:"header"`
	Query    map[string]string `json:"query" yaml:"query"`
}

//Config http路由配置结构体
type Config struct {
	name   string
	port   int
	rules  []router_http.Rule
	host   []string
	target service.IService
}

//Cert http路由驱动配置证书Cert结构体
type Cert struct {
	Key string `json:"key"`
	Crt string `json:"crt"`
}

package http_router

import (
	"github.com/eolinker/eosc"
	"github.com/eolinker/apinto/plugin"
	router_http "github.com/eolinker/apinto/router/router-http"
	"github.com/eolinker/apinto/service"
)

//DriverConfig http路由驱动配置
type DriverConfig struct {
	Driver   string                    `json:"driver" yaml:"driver"`
	Listen   int                       `json:"listen" yaml:"listen"`
	Method   []string                  `json:"method" yaml:"method"`
	Host     []string                  `json:"host" yaml:"host"`
	Rules    []DriverRule              `json:"rules" yaml:"rules"`
	Protocol string                    `json:"protocol" yaml:"protocol"`
	Cert     []Cert                    `json:"cert" yaml:"cert"`
	Target   eosc.RequireId            `json:"target" yaml:"target" skill:"github.com/eolinker/apinto/service.service.IService"`
	Plugins  map[string]*plugin.Config `json:"plugins" yaml:"plugins"`
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

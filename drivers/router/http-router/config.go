package http_router

import (
	"github.com/eolinker/eosc"
	router_http "github.com/eolinker/goku/router/router-http"
	"github.com/eolinker/goku/service"
)

//DriverConfig http路由驱动配置
type DriverConfig struct {
	Driver   string         `json:"driver" yaml:"driver"`
	Listen   int            `json:"listen" yaml:"listen"`
	Method   []string       `json:"method" yaml:"method"`
	Host     []string       `json:"host" yaml:"host"`
	Rules    []DriverRule   `json:"rules" yaml:"rules"`
	Protocol string         `json:"protocol" yaml:"protocol"`
	Cert     []Cert         `json:"cert" yaml:"cert"`
	Target   eosc.RequireId `json:"target" yaml:"target" skill:"github.com/eolinker/goku/service.service.IService"`
}

type DriverRule struct {
	Location string            `json:"location" yaml:"location"`
	Header   map[string]string `json:"header" yaml:"header"`
	Query    map[string]string `json:"query" yaml:"query"`
}

type Config struct {
	name   string
	port   int
	rules  []router_http.Rule
	host   []string
	target service.IService
}

type Cert struct {
	Key string `json:"key"`
	Crt string `json:"crt"`
}

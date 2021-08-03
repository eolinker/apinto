package http_router

import (
	"github.com/eolinker/eosc"
	router_http "github.com/eolinker/goku-eosc/router/router-http"
	"github.com/eolinker/goku-eosc/service"
)

type DriverConfig struct {
	Driver string       `json:"driver" yaml:"driver"`
	Listen int          `json:"listen" yaml:"listen"`
	Method []string     `json:"method" yaml:"method"`
	Host   []string     `json:"host" yaml:"host"`
	Rules  []DriverRule `json:"rules" yaml:"rules"`

	Target eosc.RequireId `json:"target" yaml:"target" skill:"github.com/eolinker/goku-eosc/service.service.IService"`
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

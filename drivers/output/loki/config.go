package loki

import "github.com/eolinker/eosc"

type Config struct {
	Url       string               `json:"url" yaml:"url" label:"请求地址"`
	Method    string               `json:"method" label:"请求方法" enum:"POST,PUT" default:"POST"`
	Scopes    []string             `json:"scopes" label:"作用域"`
	Headers   map[string]string    `json:"headers" yaml:"headers" label:"请求头"`
	Formatter eosc.FormatterConfig `json:"formatter" yaml:"formatter" label:"格式化配置"`
}

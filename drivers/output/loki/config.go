package loki

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/eolinker/eosc"
)

type Config struct {
	Url       string               `json:"url" yaml:"url" label:"请求地址"`
	Method    string               `json:"method" label:"请求方法" enum:"POST,PUT" default:"POST"`
	Scopes    []string             `json:"scopes" label:"作用域"`
	Headers   map[string]string    `json:"headers" yaml:"headers" label:"请求头"`
	Labels    map[string]string    `json:"labels" label:"标签"`
	Type      string               `json:"type" yaml:"type" enum:"json,line" label:"输出格式"`
	Formatter eosc.FormatterConfig `json:"formatter" yaml:"formatter" label:"格式化配置"`
}

func check(conf interface{}) (*Config, error) {
	cfg, ok := conf.(*Config)
	if !ok {
		return nil, fmt.Errorf("invalid config type: %T", conf)
	}
	if cfg.Url == "" {
		return nil, fmt.Errorf("url is empty")
	}

	u, err := url.Parse(cfg.Url)
	if err != nil {
		return nil, err
	}
	if u.Scheme == "" {
		u.Scheme = "http"
	}
	if u.Host == "" {
		return nil, fmt.Errorf("url host is empty")
	}
	cfg.Url = fmt.Sprintf("%s://%s/loki/api/v1/push", u.Scheme, u.Host)
	method := strings.ToUpper(cfg.Method)
	if method != "POST" && method != "PUT" {
		return nil, fmt.Errorf("method %s is invalid", cfg.Method)
	}
	if cfg.Type == "" {
		cfg.Type = "line"
	}
	if cfg.Labels == nil || len(cfg.Labels) == 0 {
		return nil, fmt.Errorf("labels is empty")
	}
	return cfg, nil
}

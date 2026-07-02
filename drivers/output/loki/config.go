package loki

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/eolinker/eosc"
)

type Config struct {
	Url           string               `json:"url" yaml:"url" label:"请求地址"`
	Method        string               `json:"method" label:"请求方法" enum:"POST,PUT" default:"POST"`
	Scopes        []string             `json:"scopes" label:"作用域"`
	Headers       map[string]string    `json:"headers" yaml:"headers" label:"请求头"`
	Labels        map[string]string    `json:"labels" label:"标签"`
	Type          string               `json:"type" yaml:"type" enum:"json,line" label:"输出格式"`
	ContentResize []ContentResize      `json:"content_resize" yaml:"content_resize" label:"内容截断配置" switch:"type===json"`
	Formatter     eosc.FormatterConfig `json:"formatter" yaml:"formatter" label:"格式化配置"`
}

type ContentResize struct {
	Size   int    `json:"size" label:"内容截断大小" description:"单位：字节，超过该长度的字段将被截断，0 表示不限制" minimum:"0"`
	Suffix string `json:"suffix" label:"匹配字段名后缀" description:"字段名以此后缀结尾时应用截断，例如 body 可匹配 request_body/response_body 等"`
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

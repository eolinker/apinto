package influxdb_v2

import (
	"fmt"
	"net/url"
)

type Config struct {
	Scopes      []string               `json:"scopes" label:"作用域"`
	Url         string                 `json:"url" yaml:"url" label:"请求地址"`
	Org         string                 `json:"org" yaml:"org" label:"组织名称"`
	Bucket      string                 `json:"bucket" yaml:"bucket" label:"数据桶"`
	Metrics     map[string]string      `json:"metrics" label:"标签"`
	Measurement string                 `json:"measurement" yaml:"measurement" label:"度量/表名"`
	Token       string                 `json:"token" yaml:"token" label:"访问令牌"`
	Fields      map[string]interface{} `json:"fields" yaml:"fields" label:"字段"`
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
	cfg.Url = u.String()
	
	if cfg.Token == "" {
		return nil, fmt.Errorf("token is empty")
	}
	if cfg.Org == "" {
		return nil, fmt.Errorf("org is empty")
	}
	if cfg.Bucket == "" {
		return nil, fmt.Errorf("bucket is empty")
	}
	if cfg.Measurement == "" {
		return nil, fmt.Errorf("measurement is empty")
	}
	
	return cfg, nil
}

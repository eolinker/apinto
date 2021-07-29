package nacos

import "strings"

//Config nacos驱动配置
type Config struct {
	Name   string       `json:"name"`
	Driver string       `json:"driver"`
	Scheme string       `json:"scheme"`
	Config AccessConfig `json:"config"`
}

//AccessConfig 接入地址配置
type AccessConfig struct {
	Address []string
}

func (c *Config) getScheme() string {
	scheme := strings.ToLower(c.Scheme)
	if scheme != "http" && scheme != "https" {
		scheme = "http"
	}
	return scheme
}

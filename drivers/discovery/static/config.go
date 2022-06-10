package static

import (
	"strings"
)

//Config 静态服务发现配置
type Config struct {
	Scheme   string        `json:"scheme" enum:"HTTP,HTTPS" label:"请求协议"`
	HealthOn bool          `json:"health_on" label:"健康检查"`
	Health   *HealthConfig `json:"health" switch:"health_on===true"`
}

func (c *Config) getScheme() string {

	scheme := strings.ToLower(c.Scheme)
	if scheme != "http" && scheme != "https" {
		scheme = "http"
	}
	return scheme
}

//HealthConfig 健康检查配置
type HealthConfig struct {
	Scheme      string `json:"scheme" enum:"HTTP,HTTPS" label:"请求协议"`
	Method      string `json:"method" enum:"GET,POST,PUT" label:"请求方式"`
	URL         string `json:"url" label:"请求URL"`
	SuccessCode int    `json:"success_code" label:"成功状态码" minimum:"99"`
	Period      int    `json:"period" label:"检查频率（单位：s）" minimum:"1" default:"30"`
	Timeout     int    `json:"timeout" label:"超时时间（单位：ms）"`
}

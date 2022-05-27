package static

import (
	"strings"
)

//Config 静态服务发现配置
type Config struct {
	Scheme   string        `json:"scheme" enum:"HTTP,HTTPS"`
	HealthOn bool          `json:"health_on"`
	Health   *HealthConfig `json:"health" switch:"health_on=ture"`
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
	Scheme      string `json:"scheme"`
	Method      string `json:"method" enum:"GET,POST,PUT"`
	URL         string `json:"url"`
	SuccessCode int    `json:"success_code"`
	Period      int    `json:"period"`
	Timeout     int    `json:"timeout"`
}

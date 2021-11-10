package static

import "strings"

//Config 静态服务发现配置
type Config struct {
	//Name     string            `json:"factoryName"`
	//Driver   string            `json:"driver"`
	//Labels   map[string]string `json:"labels"`
	Scheme   string        `json:"scheme"`
	Health   *HealthConfig `json:"health"`
	HealthOn bool          `json:"health_on"`
}

func (c *Config) getScheme() string {
	scheme := strings.ToLower(c.Scheme)
	if scheme != "http-service" && scheme != "https" {
		scheme = "http-service"
	}
	return scheme
}

//HealthConfig 健康检查配置
type HealthConfig struct {
	Scheme      string `json:"scheme"`
	Method      string `json:"method"`
	URL         string `json:"url"`
	SuccessCode int    `json:"success_code"`
	Period      int    `json:"period"`
	Timeout     int    `json:"timeout"`
}

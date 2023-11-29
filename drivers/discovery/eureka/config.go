package eureka

import (
	"fmt"
	"net/url"
	"strings"
)

const defaultScheme = "http"

// Config eureka驱动配置
type Config struct {
	Config AccessConfig `json:"config" label:"配置信息"`
}

// AccessConfig 接入地址配置
type AccessConfig struct {
	Address []string          `json:"address" label:"eureka地址"`
	Params  map[string]string `json:"params" label:"参数"`
}

func (c *Config) getAddress() []string {
	adds := make([]string, 0, len(c.Config.Address))
	for _, a := range c.Config.Address {
		if !strings.HasPrefix(a, "http://") && !strings.HasPrefix(a, "https://") {
			a = fmt.Sprintf("%s://%s", defaultScheme, a)
		}
		adds = append(adds, a)
	}
	return adds
}
func (c *Config) getParams() url.Values {
	p := url.Values{}
	for k, v := range c.Config.Params {
		p.Set(k, v)
	}
	return p
}

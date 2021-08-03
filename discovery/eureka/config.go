package eureka

import (
	"fmt"
	"net/url"
	"strings"
)

//Config eureka驱动配置
type Config struct {
	Name   string       `json:"name"`
	Driver string       `json:"driver"`
	Scheme string       `json:"scheme"`
	Config AccessConfig `json:"config"`
}

//AccessConfig 接入地址配置
type AccessConfig struct {
	Address []string
	Params  map[string]string
}

func (c *Config) getScheme() string {
	scheme := strings.ToLower(c.Scheme)
	if scheme != "http" && scheme != "https" {
		scheme = "http"
	}
	return scheme
}

func (c *Config) getAddress() []string {
	scheme := c.getScheme()
	adds := make([]string, 0, len(c.Config.Address))
	for _, a := range c.Config.Address {
		if !strings.HasPrefix(a, "http://") && !strings.HasPrefix(a, "https://") {
			a = fmt.Sprintf("%s://%s", scheme, a)
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

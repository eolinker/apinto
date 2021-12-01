package consul

import (
	"strings"

	"github.com/hashicorp/consul/api"
)

//Config consul驱动配置
type Config struct {
	Name   string       `json:"name"`
	Driver string       `json:"driver"`
	Scheme string       `json:"scheme"`
	Config AccessConfig `json:"config"`
}

//AccessConfig 接入地址配置
type AccessConfig struct {
	Address []string          `json:"address"`
	Params  map[string]string `json:"params"`
}

type consulClients struct {
	clients []*api.Client
}

func (c *Config) getScheme() string {
	scheme := strings.ToLower(c.Scheme)
	if scheme != "http" && scheme != "https" {
		scheme = "http"
	}
	return scheme
}

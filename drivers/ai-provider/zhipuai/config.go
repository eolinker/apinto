package zhipuai

import (
	"fmt"
	"github.com/eolinker/eosc"
)

type Config struct {
	APIKey       string `json:"zhipuai_api_key"`
	Organization string `json:"zhipuai_organization"`
}

func checkConfig(v interface{}) (*Config, error) {
	conf, ok := v.(*Config)
	if !ok {
		return nil, eosc.ErrorConfigType
	}
	if conf.APIKey == "" {
		return nil, fmt.Errorf("api_key is required")
	}
	return conf, nil
}

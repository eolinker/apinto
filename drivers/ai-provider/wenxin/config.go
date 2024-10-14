package wenxin

import (
	"fmt"

	"github.com/eolinker/eosc"
)

type Config struct {
	APIKey    string `json:"api_key"`
	SecretKey string `json:"secret_key"`
}

func checkConfig(v interface{}) (*Config, error) {
	conf, ok := v.(*Config)
	if !ok {
		return nil, eosc.ErrorConfigType
	}
	if conf.APIKey == "" {
		return nil, fmt.Errorf("api_key is required")
	}
	if conf.SecretKey == "" {
		return nil, fmt.Errorf("secret_key is required")
	}
	return conf, nil
}

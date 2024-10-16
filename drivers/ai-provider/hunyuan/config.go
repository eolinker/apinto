package hunyuan

import (
	"fmt"

	"github.com/eolinker/eosc"
)

type Config struct {
	SecretID  string `json:"secret_id"`
	SecretKey string `json:"secret_key"`
}

func checkConfig(v interface{}) (*Config, error) {
	conf, ok := v.(*Config)
	if !ok {
		return nil, eosc.ErrorConfigType
	}
	if conf.SecretID == "" {
		return nil, fmt.Errorf("secret_id is required")
	}
	if conf.SecretKey == "" {
		return nil, fmt.Errorf("secret_key is required")
	}
	return conf, nil
}

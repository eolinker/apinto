package spark

import (
	"fmt"

	"github.com/eolinker/eosc"
)

type Config struct {
	APIPassword string `json:"api_password"`
}

func checkConfig(v interface{}) (*Config, error) {
	conf, ok := v.(*Config)
	if !ok {
		return nil, eosc.ErrorConfigType
	}
	if conf.APIPassword == "" {
		return nil, fmt.Errorf("api_password is required")
	}
	return conf, nil
}

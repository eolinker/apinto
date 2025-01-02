package ai_balance

import "github.com/eolinker/eosc"

type Config struct {
}

func checkConfig(v interface{}) (*Config, error) {
	conf, ok := v.(*Config)
	if !ok {
		return nil, eosc.ErrorConfigType
	}
	return conf, nil
}

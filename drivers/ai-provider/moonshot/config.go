package moonshot

import (
	"fmt"

	"github.com/eolinker/eosc"
)

type Config struct {
	APIKey string `json:"moonshot_api_key"`
	Base   string `json:"base"`
}

func checkConfig(v interface{}) (*Config, error) {
	conf, ok := v.(*Config)
	if !ok {
		return nil, eosc.ErrorConfigType
	}
	if conf.APIKey == "" {
		return nil, fmt.Errorf("moonshot_api_key is required")
	}
	return conf, nil
}

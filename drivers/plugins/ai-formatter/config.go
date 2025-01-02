package ai_formatter

import (
	"github.com/eolinker/eosc"
)

type Config struct {
	Provider eosc.RequireId `json:"provider" skill:"github.com/eolinker/apinto/convert.key.IKeyPool"`
	Model    string         `json:"model"`
	Config   string         `json:"config"`
}

func checkConfig(v interface{}) (*Config, error) {
	conf, ok := v.(*Config)
	if !ok {
		return nil, eosc.ErrorConfigType
	}
	return conf, nil
}

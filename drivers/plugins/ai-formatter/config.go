package ai_formatter

import (
	ai_convert "github.com/eolinker/apinto/ai-convert"
	"github.com/eolinker/eosc"
)

type Config struct {
	Provider  string `json:"provider"`
	Model     string `json:"model"`
	ModelType string `json:"model_type"`
	Config    string `json:"config"`
}

func checkConfig(v interface{}) (*Config, error) {
	conf, ok := v.(*Config)
	if !ok {
		return nil, eosc.ErrorConfigType
	}

	if conf.ModelType == "" {
		conf.ModelType = ai_convert.ModelTypeOpenAIChat.String()
	}

	return conf, nil
}

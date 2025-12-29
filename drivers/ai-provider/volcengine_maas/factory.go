package openAI

import (
	"encoding/json"
	"github.com/eolinker/eosc"

	ai_convert "github.com/eolinker/apinto/ai-convert"
)

const (
	provider = "volcengine_mass"
)

var (
	driverCreate = eosc.BuildUntyped[ai_convert.ModelType, ai_convert.IConvertDriverCreateFunc[Config]]()
)

func init() {
	ai_convert.RegisterConverterCreateFunc(provider, Create)
}

func Create(cfg string) (ai_convert.IConverter, error) {
	var conf Config
	err := json.Unmarshal([]byte(cfg), &conf)
	if err != nil {
		return nil, err
	}
	err = checkConfig(&conf)
	if err != nil {
		return nil, err
	}
	if conf.BaseUrl == "" {
		conf.BaseUrl = "https://ark.cn-beijing.volces.com/api/v3"
	}
	return ai_convert.NewConverter(provider, &conf, driverCreate.List())
}

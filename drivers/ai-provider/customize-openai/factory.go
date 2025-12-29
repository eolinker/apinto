package customize_openai

import (
	"encoding/json"
	"github.com/eolinker/eosc"

	ai_convert "github.com/eolinker/apinto/ai-convert"
)

const (
	provider = "customize-openai"
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
	return ai_convert.NewConverter(provider, &conf, driverCreate.List())
}

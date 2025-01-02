package convert

import "github.com/eolinker/eosc/eocontext"

type IConverterFactory interface {
	Create(cfg string) (IConverterDriver, error)
}

type IConverterDriver interface {
	GetModel(model string) (FGenerateConfig, bool)
	GetConverter(model string) (IConverter, bool)
}

type IConverter interface {
	RequestConvert(ctx eocontext.EoContext, extender map[string]interface{}) error
	ResponseConvert(ctx eocontext.EoContext) error
}

type IChildConverter interface {
	IConverter
	Endpoint() string
}
type FGenerateConfig func(cfg string) (map[string]interface{}, error)

func CheckSkill(skill string) bool {
	return skill == "github.com/eolinker/apinto/convert.key.IKeyPool"
}

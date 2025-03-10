package ai_convert

import "github.com/eolinker/eosc/eocontext"

type IConverterFactory interface {
	Create(cfg string) (IConverter, error)
}

type IConverterCreateFunc func(cfg string) (IConverter, error)

//type IConverterDriver interface {
//	GetModel(model string) (FGenerateConfig, bool)
//	GetConverter(model string) (IConverter, bool)
//}

type IConverter interface {
	RequestConvert(ctx eocontext.EoContext, extender map[string]interface{}) error
	ResponseConvert(ctx eocontext.EoContext) error
}

type IChildConverter interface {
	IConverter
	Endpoint() string
}
type FGenerateConfig func(cfg string) (map[string]interface{}, error)

func CheckKeySourceSkill(skill string) bool {
	return skill == "github.com/eolinker/apinto/convert.key.IKeyResource"
}

func CheckProviderSkill(skill string) bool {
	return skill == "github.com/eolinker/apinto/convert.provider.IProvider"
}

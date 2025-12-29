package ai_convert

import (
	"fmt"
	"github.com/eolinker/eosc"
	"github.com/eolinker/eosc/eocontext"
)

type IConverterCreateFunc func(cfg string) (IConverter, error)

type IConvertDriverCreateFunc[T any] func(*T) (IConverterDriver, error)

type ModelType string

func (m ModelType) String() string {
	return string(m)
}

const (
	ModelTypeChat  ModelType = "chat"
	ModelTypeImage ModelType = "image"
)

type IConverterFactory interface {
	Create(cfg string) (IConverter, error)
}

type IConverter interface {
	Get(modelType ModelType) (IConverterDriver, bool)
	ModelTypeList() []ModelType
}

type IConverterDriver interface {
	Provider() string
	ModelType() ModelType
	RequestConvert(ctx eocontext.EoContext, extender map[string]interface{}) error
	ResponseConvert(ctx eocontext.EoContext) error
}

type IChildConverter interface {
	IConverterDriver
	Endpoint() string
}
type FGenerateConfig func(cfg string) (map[string]interface{}, error)

func CheckKeySourceSkill(skill string) bool {
	return skill == "github.com/eolinker/apinto/convert.key.IKeyResource"
}

func CheckProviderSkill(skill string) bool {
	return skill == "github.com/eolinker/apinto/convert.provider.IProvider"
}

func NewConverter[T any](provider string, cfg *T, fns []IConvertDriverCreateFunc[T]) (IConverter, error) {
	if len(fns) == 0 {
		return nil, fmt.Errorf("no driver found for %s provider", provider)
	}
	c := &Converter{
		drivers: eosc.BuildUntyped[ModelType, IConverterDriver](),
	}
	for _, fn := range fns {
		driver, err := fn(cfg)
		if err != nil {
			return nil, err
		}
		c.drivers.Set(driver.ModelType(), driver)
	}

	return c, nil
}

type Converter struct {
	drivers eosc.Untyped[ModelType, IConverterDriver]
}

func (c *Converter) Get(modelType ModelType) (IConverterDriver, bool) {
	if c.drivers == nil {
		return nil, false
	}
	return c.drivers.Get(modelType)
}

func (c *Converter) ModelTypeList() []ModelType {
	return c.drivers.Keys()
}

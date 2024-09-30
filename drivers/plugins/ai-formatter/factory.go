package ai_formatter

import (
	"github.com/eolinker/apinto/drivers"
	"github.com/eolinker/eosc"
)

const (
	Name = "ai_formatter"
)

func Register(register eosc.IExtenderDriverRegister) {
	register.RegisterExtenderDriver(Name, NewFactory())
}

type Factory struct {
	eosc.IExtenderDriverFactory
}

func NewFactory() *Factory {
	return &Factory{
		IExtenderDriverFactory: drivers.NewFactory[Config](Create),
	}
}

func (f *Factory) Create(profession string, name string, label string, desc string, params map[string]interface{}) (eosc.IExtenderDriver, error) {
	return f.IExtenderDriverFactory.Create(profession, name, label, desc, params)
}

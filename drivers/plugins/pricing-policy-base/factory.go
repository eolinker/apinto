package pricing_policy_base

import (
	"github.com/eolinker/apinto/drivers"
	"github.com/eolinker/eosc"
)

// API发布时，绑定该插件，实现AI代理功能

const (
	Name = "pricing_policy_base"
)

func Register(register eosc.IExtenderDriverRegister) {
	register.RegisterExtenderDriver(Name, NewFactory())
}

type Factory struct {
	eosc.IExtenderDriverFactory
}

func NewFactory() *Factory {
	return &Factory{
		IExtenderDriverFactory: drivers.NewFactory[Config](Create, check),
	}
}

func (f *Factory) Create(profession string, name string, label string, desc string, params map[string]interface{}) (eosc.IExtenderDriver, error) {
	return f.IExtenderDriverFactory.Create(profession, name, label, desc, params)
}

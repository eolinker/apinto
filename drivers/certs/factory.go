package certs

import (
	"github.com/eolinker/eosc"

	"github.com/eolinker/apinto/drivers"
)

func Register(register eosc.IExtenderDriverRegister) {
	_ = register.RegisterExtenderDriver("ssl-server", newFactory())
	//setting.RegisterSetting("ssl-server", controller)
}

func newFactory() eosc.IExtenderDriverFactory {
	return &factory{IExtenderDriverFactory: drivers.NewFactory[Config](Create)}
}

type factory struct {
	eosc.IExtenderDriverFactory
}

func (f *factory) Create(profession string, name string, label string, desc string, params map[string]interface{}) (eosc.IExtenderDriver, error) {
	controller.driver = name
	controller.profession = profession
	return f.IExtenderDriverFactory.Create(profession, name, label, desc, params)
}

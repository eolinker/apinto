package app

import (
	"sync"

	"github.com/eolinker/eosc"
	"github.com/eolinker/eosc/common/bean"

	"github.com/eolinker/apinto/drivers"
	"github.com/eolinker/apinto/drivers/app/manager"
)

const (
	Name = "plugin_app"
)

var (
	ones       sync.Once
	appManager manager.IManager
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
	ones.Do(func() {
		bean.Autowired(&appManager)
	})

	return f.IExtenderDriverFactory.Create(profession, name, label, desc, params)
}

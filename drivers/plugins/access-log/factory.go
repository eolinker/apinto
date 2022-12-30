package access_log

import (
	"sync"

	scope_manager "github.com/eolinker/apinto/drivers/scope-manager"

	"github.com/eolinker/apinto/drivers"
	"github.com/eolinker/eosc"
	"github.com/eolinker/eosc/common/bean"
)

const (
	Name = "access_log"
)

var (
	workers      eosc.IWorkers
	scopeManager scope_manager.IManager
	once         sync.Once
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
	once.Do(func() {
		bean.Autowired(&workers)
		bean.Autowired(&scopeManager)
	})

	return f.IExtenderDriverFactory.Create(profession, name, label, desc, params)
}

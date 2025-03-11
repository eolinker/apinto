package ai_prompt

import (
	"sync"

	"github.com/eolinker/apinto/drivers"
	"github.com/eolinker/eosc"
	"github.com/eolinker/eosc/common/bean"
)

const (
	Name = "ai_prompt"
)

var (
	customerVar eosc.ICustomerVar
	once        sync.Once
)

func init() {
	once.Do(func() {
		bean.Autowired(&customerVar)
	})
}

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

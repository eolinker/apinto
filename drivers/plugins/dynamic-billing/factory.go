package dynamic_billing

import (
	"github.com/eolinker/apinto/price-calcular"
	"sync"

	"github.com/eolinker/apinto/drivers"
	"github.com/eolinker/eosc"
	"github.com/eolinker/eosc/common/bean"
)

const (
	Name = "dynamic_billing"
)

var (
	policyManager price_calcular.IManager
	customerVar   eosc.ICustomerVar
	once          sync.Once
)

// Register 注册 extender 驱动
func Register(register eosc.IExtenderDriverRegister) {
	register.RegisterExtenderDriver(Name, NewFactory())
}

type Factory struct {
	eosc.IExtenderDriverFactory
}

func NewFactory() *Factory {
	return &Factory{
		IExtenderDriverFactory: drivers.NewFactory[Config](Create, checkConfig),
	}
}

func (f *Factory) Create(profession string, name string, label string, desc string, params map[string]interface{}) (eosc.IExtenderDriver, error) {
	once.Do(func() {
		bean.Autowired(&policyManager)
		bean.Autowired(&customerVar)
	})

	return f.IExtenderDriverFactory.Create(profession, name, label, desc, params)
}

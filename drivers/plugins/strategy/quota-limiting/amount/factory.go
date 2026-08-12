package amount

import (
	"github.com/eolinker/apinto/drivers"
	"github.com/eolinker/eosc"
	"github.com/eolinker/eosc/common/bean"
	"sync"
)

const (
	Name = "strategy-plugin-quota-limiting-amount"
)

var (
	customerVar eosc.ICustomerVar
	once        sync.Once
)

func Register(register eosc.IExtenderDriverRegister) {
	_ = register.RegisterExtenderDriver(Name, NewFactory())
}

func NewFactory() eosc.IExtenderDriverFactory {
	once.Do(func() {
		bean.Autowired(&customerVar)
	})
	return drivers.NewFactory[Config](Create, CheckConfig)
}

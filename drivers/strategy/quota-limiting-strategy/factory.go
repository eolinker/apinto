package quota_limiting_strategy

import (
	"github.com/eolinker/apinto/drivers/strategy"
	"github.com/eolinker/eosc/common/bean"
	"github.com/eolinker/eosc/setting"
	"reflect"
	"sync"

	"github.com/eolinker/apinto/drivers"
	"github.com/eolinker/eosc"
)

const Name = "strategy-quota-limiting"

var (
	configType  = reflect.TypeOf((*Config)(nil))
	customerVar eosc.ICustomerVar
	controller  *strategy.Controller
	once        sync.Once
)

// Register 注册配额策略扩展驱动与 Setting Controller
func Register(register eosc.IExtenderDriverRegister) {
	register.RegisterExtenderDriver(Name, newFactory())
}

type factory struct {
	eosc.IExtenderDriverFactory
}

func newFactory() eosc.IExtenderDriverFactory {
	return &factory{
		IExtenderDriverFactory: drivers.NewFactory[Config](Create, Check),
	}
}

func (f *factory) Create(profession string, name string, label string, desc string, params map[string]interface{}) (eosc.IExtenderDriver, error) {
	controller = strategy.NewController(profession, name, configType)
	once.Do(func() {
		bean.Autowired(&customerVar)
		setting.RegisterSetting("strategies-quota-limiting", controller)
	})

	return f.IExtenderDriverFactory.Create(profession, name, label, desc, params)
}

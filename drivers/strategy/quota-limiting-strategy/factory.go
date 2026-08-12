package quota_limiting_strategy

import (
	"github.com/eolinker/apinto/drivers/strategy"
	"reflect"
	"sync"
	
	"github.com/eolinker/apinto/drivers"
	"github.com/eolinker/eosc"
	"github.com/eolinker/eosc/common/bean"
	"github.com/eolinker/eosc/setting"
)

const Name = "strategy-quota-limiting"

var (
	configType  = reflect.TypeOf((*Config)(nil))
	customerVar eosc.ICustomerVar
	controller  *strategy.Controller
	once        sync.Once
)

func init() {
	once.Do(func() {
		bean.Autowired(&customerVar)
	})
}

// Register 注册配额策略扩展驱动与 Setting Controller
func Register(register eosc.IExtenderDriverRegister) {
	register.RegisterExtenderDriver(Name, newFactory())
	setting.RegisterSetting("strategies-quota-limiting", controller)
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
	return f.IExtenderDriverFactory.Create(profession, name, label, desc, params)
}

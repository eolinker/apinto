package failover_strategy

import (
	"github.com/eolinker/apinto/drivers/strategy"
	"reflect"
	"sync"

	"github.com/eolinker/apinto/drivers"
	"github.com/eolinker/eosc"
	"github.com/eolinker/eosc/setting"
)

const Name = "strategy-failover"

var (
	configType = reflect.TypeOf((*Config)(nil))
	controller *strategy.Controller
	once       = sync.Once{}
)

// Register 注册灾备策略驱动工厂
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
		setting.RegisterSetting("strategies-failover", controller)
	})
	return f.IExtenderDriverFactory.Create(profession, name, label, desc, params)
}

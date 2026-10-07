package model_mapping

import (
	"reflect"
	"sync"

	"github.com/eolinker/apinto/drivers"
	"github.com/eolinker/apinto/drivers/strategy"
	"github.com/eolinker/eosc"
	"github.com/eolinker/eosc/setting"
)

const Name = "strategy-model-mapping"

var (
	configType = reflect.TypeOf((*Config)(nil))
	controller *strategy.Controller
	once       = sync.Once{}
)

// Register 注册模型映射策略驱动工厂
func Register(register eosc.IExtenderDriverRegister) {
	register.RegisterExtenderDriver(Name, newFactory())
}

type factory struct {
	eosc.IExtenderDriverFactory
}

func newFactory() eosc.IExtenderDriverFactory {
	return &factory{
		IExtenderDriverFactory: drivers.NewFactory[Config](Create, CheckConfig),
	}
}

func (f *factory) Create(profession string, name string, label string, desc string, params map[string]interface{}) (eosc.IExtenderDriver, error) {
	controller = strategy.NewController(profession, name, configType)
	once.Do(func() {
		setting.RegisterSetting("strategies-model-mapping", controller)
	})
	return f.IExtenderDriverFactory.Create(profession, name, label, desc, params)
}

package data_mask_strategy

import (
	"reflect"
	"sync"

	"github.com/eolinker/apinto/drivers/strategy/data-mask-strategy/mask/inner"
	json_path "github.com/eolinker/apinto/drivers/strategy/data-mask-strategy/mask/json-path"
	"github.com/eolinker/apinto/drivers/strategy/data-mask-strategy/mask/keyword"
	"github.com/eolinker/apinto/drivers/strategy/data-mask-strategy/mask/regex"

	"github.com/eolinker/apinto/drivers"
	"github.com/eolinker/eosc"
	"github.com/eolinker/eosc/setting"
)

const Name = "strategy-data_mask"

var (
	configType = reflect.TypeOf((*Config)(nil))
	once       sync.Once
)

// Register 注册http路由驱动工厂
func Register(register eosc.IExtenderDriverRegister) {
	register.RegisterExtenderDriver(Name, newFactory())
	setting.RegisterSetting("strategies-data_mask", controller)
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
	once.Do(func() {
		inner.Register()
		json_path.Register()
		keyword.Register()
		regex.Register()
	})
	controller.driver = name
	controller.profession = profession
	return f.IExtenderDriverFactory.Create(profession, name, label, desc, params)
}

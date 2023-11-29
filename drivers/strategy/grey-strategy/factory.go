package grey_strategy

import (
	"reflect"

	"github.com/eolinker/eosc"
	"github.com/eolinker/eosc/setting"

	"github.com/eolinker/apinto/drivers"
)

const Name = "strategy-grey"

var (
	configType = reflect.TypeOf((*Config)(nil))
)

// Register 注册http路由驱动工厂
func Register(register eosc.IExtenderDriverRegister) {

	_ = register.RegisterExtenderDriver(Name, newFactory())

	err := setting.RegisterSetting("strategies-grey", controller)
	if err != nil {
		return
	}
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
	controller.driver = name
	controller.profession = profession
	return f.IExtenderDriverFactory.Create(profession, name, label, desc, params)
}

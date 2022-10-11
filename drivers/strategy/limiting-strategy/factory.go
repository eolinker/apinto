package limiting_strategy

import (
	"github.com/eolinker/eosc"
	"github.com/eolinker/eosc/setting"
	"github.com/eolinker/eosc/utils/schema"
	"reflect"
)

const Name = "strategy-limiting"

var (
	configType = reflect.TypeOf((*Config)(nil))
)

//Register 注册http路由驱动工厂
func Register(register eosc.IExtenderDriverRegister) {

	register.RegisterExtenderDriver(Name, newFactory())
	setting.RegisterSetting("strategies-limiting", controller)
}

type factory struct {
	render interface{}
}

func newFactory() *factory {
	render, err := schema.Generate(configType, nil)
	if err != nil {
		panic(err)
	}
	return &factory{
		render: render,
	}
}

func (f *factory) Render() interface{} {
	return f.render
}

func (f *factory) Create(profession string, name string, label string, desc string, params map[string]interface{}) (eosc.IExtenderDriver, error) {
	controller.driver = name
	controller.profession = profession
	return &driver{}, nil
}

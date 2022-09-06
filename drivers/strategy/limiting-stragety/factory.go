package limiting_stragety

import (
	"github.com/eolinker/eosc"
	"reflect"
)

const Name = "strategy-limiting"

//Register 注册http路由驱动工厂
func Register(register eosc.IExtenderDriverRegister) {
	register.RegisterExtenderDriver(Name, newFactory())
}

type factory struct {
	configType reflect.Type
	render     interface{}
}

func newFactory() *factory {
	return &factory{}
}

func (f *factory) Render() interface{} {
	return f.render
}

func (f *factory) Create(profession string, name string, label string, desc string, params map[string]interface{}) (eosc.IExtenderDriver, error) {
	return &driver{configType: f.configType}, nil
}

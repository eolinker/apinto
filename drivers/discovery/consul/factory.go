package consul

import (
	"github.com/eolinker/eosc"
	"reflect"
)

var name = "discovery_consul"

//Register 注册consul驱动工厂
func Register(register eosc.IExtenderDriverRegister) {
	register.RegisterExtenderDriver(name, NewFactory())
}

type factory struct {
}

//NewFactory 创建consul驱动工厂
func NewFactory() eosc.IExtenderDriverFactory {
	return &factory{}
}

//Create 创建consul驱动
func (f *factory) Create(profession string, name string, label string, desc string, params map[string]interface{}) (eosc.IExtenderDriver, error) {
	return &driver{
		profession: profession,
		name:       name,
		label:      label,
		desc:       desc,
		driver:     driverName,
		configType: reflect.TypeOf((*Config)(nil)),
	}, nil
}

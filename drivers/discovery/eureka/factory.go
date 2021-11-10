package eureka

import (
	"reflect"

	"github.com/eolinker/eosc"
)

var name = "discovery_eureka"

//Register 注册eureka驱动工厂
func Register(register eosc.IExtenderDriverRegister) {
	register.RegisterExtenderDriver(name, NewFactory())
}

type factory struct {
	profession string
	name       string
	label      string
	desc       string
	params     map[string]string
}

//NewFactory 创建eureka驱动工厂
func NewFactory() eosc.IExtenderDriverFactory {
	return &factory{}
}

//Create 创建eureka驱动
func (f *factory) Create(profession string, name string, label string, desc string, params map[string]string) (eosc.IExtenderDriver, error) {
	return &driver{
		profession: profession,
		name:       name,
		label:      label,
		desc:       desc,
		driver:     driverName,
		configType: reflect.TypeOf((*Config)(nil)),
		params:     params,
	}, nil
}

package static

import (
	"reflect"

	"github.com/eolinker/eosc"
)

//Register 注册静态服务发现的驱动工厂
func Register(register eosc.IExtenderRegister) {
	register.RegisterExtender("eolinker:goku:discovery_static", NewFactory())
}

type factory struct {
	profession string
	name       string
	label      string
	desc       string
	params     map[string]string
}

//NewFactory 创建静态服务发现的驱动工厂
func NewFactory() eosc.IExtenderDriverFactory {
	return &factory{}
}

//Create 创建静态服务发现驱动
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

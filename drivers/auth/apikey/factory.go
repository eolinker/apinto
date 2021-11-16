package apikey

import (
	"reflect"

	"github.com/eolinker/eosc"
)

var name = "auth_apikey"

//Register 注册auth驱动工厂
func Register(register eosc.IExtenderDriverRegister) {
	register.RegisterExtenderDriver(name, NewFactory())
}

type factory struct {
}

//Create 创建apikey驱动
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

//NewFactory 生成一个 auth_apiKey工厂
func NewFactory() eosc.IExtenderDriverFactory {
	return &factory{}
}

package jwt

import (
	"reflect"

	"github.com/eolinker/eosc"
)

//Register 注册jwt鉴权驱动工厂
func Register(register eosc.IExtenderRegister) {
	register.RegisterExtender("auth_jwt", NewFactory())
}

type factory struct {
	profession string
	name       string
	label      string
	desc       string
	params     map[string]string
}

//NewFactory 创建jwt鉴权驱动工厂
func NewFactory() eosc.IExtenderDriverFactory {
	return &factory{}
}

//Create 创建jwt鉴权驱动
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

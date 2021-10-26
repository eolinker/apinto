package syslog

import (
	"reflect"

	"github.com/eolinker/eosc"
)

//Register 注册syslog驱动工厂
func Register(register eosc.IExtenderRegister) {
	register.RegisterExtender("log_syslog", NewFactory())
}

type factory struct {
}

//NewFactory 创建syslog驱动工厂
func NewFactory() eosc.IExtenderDriverFactory {
	return &factory{}
}

//Create 创建syslog驱动
func (f *factory) Create(profession string, name string, label string, desc string, params map[string]string) (eosc.IExtenderDriver, error) {
	return &driver{
		profession: profession,
		name:       name,
		label:      label,
		desc:       desc,
		driver:     driverName,
		configType: reflect.TypeOf((*DriverConfig)(nil)),
		params:     params,
	}, nil
}

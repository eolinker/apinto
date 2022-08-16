package syslog

import (
	"github.com/eolinker/eosc"
	"reflect"
)

const name = "syslog_output"

//Register 注册file_output驱动工厂
func Register(register eosc.IExtenderDriverRegister) {
	register.RegisterExtenderDriver(name, NewFactory())
}

type Factory struct {
}

func NewFactory() *Factory {
	return &Factory{}
}

func (f *Factory) Create(profession string, name string, label string, desc string, params map[string]interface{}) (eosc.IExtenderDriver, error) {
	return &Driver{
		configType: reflect.TypeOf((*Config)(nil)),
	}, nil
}

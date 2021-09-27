package stdlog

import (
	"reflect"

	"github.com/eolinker/eosc"
)

//Register 注册stdlog驱动工厂
func Register() {
	eosc.DefaultProfessionDriverRegister.RegisterProfessionDriver("eolinker:goku:log_stdlog", NewFactory())
}

type factory struct {
}

//NewFactory 创建stdlog驱动工厂
func NewFactory() eosc.IProfessionDriverFactory {
	return &factory{}
}

//Create 创建stdlog驱动
func (f *factory) Create(profession string, name string, label string, desc string, params map[string]string) (eosc.IProfessionDriver, error) {
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

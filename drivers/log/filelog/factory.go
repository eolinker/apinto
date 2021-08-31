package filelog

import (
	"github.com/eolinker/eosc"
	"reflect"
)

//Register 注册filelog驱动工厂
func Register() {
	eosc.DefaultProfessionDriverRegister.RegisterProfessionDriver("eolinker:goku:log_filelog", NewFactory())
}

type factory struct {
}

//NewFactory 创建filelog驱动工厂
func NewFactory() eosc.IProfessionDriverFactory {
	return &factory{}
}

//ExtendInfo 返回filelog驱动工厂的信息
func (f *factory) ExtendInfo() eosc.ExtendInfo {
	return eosc.ExtendInfo{
		ID:      "eolinker:goku:log_filelog",
		Group:   "eolinker",
		Project: "goku",
		Name:    "filelog",
	}
}

//Create 创建filelog驱动
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

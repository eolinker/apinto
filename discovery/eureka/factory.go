package eureka

import (
	"github.com/eolinker/eosc"
	"reflect"
)

//Register 注册eureka驱动工厂
func Register() {
	eosc.DefaultProfessionDriverRegister.RegisterProfessionDriver("eolinker:goku:discovery_eureka", NewFactory())
}

type factory struct {
	profession string
	name       string
	label      string
	desc       string
	params     map[string]string
}

//NewFactory 创建eureka驱动工厂
func NewFactory() eosc.IProfessionDriverFactory {
	return &factory{}
}

//ExtendInfo 返回eureka驱动工厂信息
func (f *factory) ExtendInfo() eosc.ExtendInfo {
	return eosc.ExtendInfo{
		ID:      "eolinker:goku:discover_eureka",
		Group:   "eolinker",
		Project: "goku",
		Name:    "eureka",
	}
}

//Create 创建eureka驱动
func (f *factory) Create(profession string, name string, label string, desc string, params map[string]string) (eosc.IProfessionDriver, error) {
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

package upstream_http_anonymous

import (
	"reflect"

	"github.com/eolinker/eosc"
)

//Register 注册匿名http_proxy驱动工厂
func Register() {
	eosc.DefaultProfessionDriverRegister.RegisterProfessionDriver("eolinker:goku:http_proxy", NewFactory())
}

type factory struct {
	profession string
	name       string
	label      string
	desc       string
	params     map[string]string
}

//NewFactory 创建匿名http_proxy驱动工厂
func NewFactory() eosc.IProfessionDriverFactory {
	return &factory{}
}

//ExtendInfo 返回匿名http_proxy驱动工厂的信息
func (f *factory) ExtendInfo() eosc.ExtendInfo {
	return eosc.ExtendInfo{
		ID:      "eolinker:goku:upstream_http_proxy",
		Group:   "eolinker",
		Project: "goku",
		Name:    "http_proxy",
	}
}

//Create 创建匿名http_proxy驱动
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

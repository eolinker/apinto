package apikey

import (
	"github.com/eolinker/eosc"
	"reflect"
)

//Register 注册auth驱动工厂
func Register() {
	eosc.DefaultProfessionDriverRegister.RegisterProfessionDriver("eolinker:goku:auth_apikey", NewFactory())
}

type factory struct {
	profession string
	name       string
	label      string
	desc       string
	params     map[string]string
}

//ExtendInfo 返回auth_apikey的驱动工厂信息
func (f *factory) ExtendInfo() eosc.ExtendInfo {
	return eosc.ExtendInfo{
		ID:      "eolinker:goku:auth_apikey",
		Group:   "eolinker",
		Project: "goku",
		Name:    "apikey",
	}
}

//Create 创建apikey驱动
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

//NewFactory 生成一个 auth_apiKey工厂
func NewFactory() eosc.IProfessionDriverFactory {
	return &factory{}
}

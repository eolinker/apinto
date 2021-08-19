package httplog

import (
	"github.com/eolinker/eosc"
	"reflect"
)

//Register 注册httplog驱动工厂
func Register() {
	eosc.DefaultProfessionDriverRegister.RegisterProfessionDriver("eolinker:goku:log_httplog", NewFactory())
}

type factory struct {
}

//NewFactory 创建httplog驱动工厂
func NewFactory() eosc.IProfessionDriverFactory {
	return &factory{}
}

//ExtendInfo 返回httplog驱动工厂的信息
func (f *factory) ExtendInfo() eosc.ExtendInfo {
	return eosc.ExtendInfo{
		ID:      "eolinker:goku:log_httplog",
		Group:   "eolinker",
		Project: "goku",
		Name:    "httplog",
	}
}

//Create 创建httplog驱动
func (f *factory) Create(profession string, name string, label string, desc string, params map[string]string) (eosc.IProfessionDriver, error) {
	//if o, has := params["access_log"]; has && o == "true" {
	//	return &accessDriver{
	//		profession: profession,
	//		name:       name,
	//		label:      label,
	//		desc:       desc,
	//		driver:     driverName,
	//		configType: reflect.TypeOf((*DriverConfigAccess)(nil)),
	//		params:     params,
	//	}, nil
	//}
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

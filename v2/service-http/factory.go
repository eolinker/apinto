package service_http

import (
	round_robin "github.com/eolinker/apinto/upstream/round-robin"
	"reflect"

	"github.com/eolinker/apinto/drivers/discovery/static"
	"github.com/eolinker/eosc"
)

var DriverName = "service_http"
var (
	defaultHttpDiscovery = static.CreateAnonymous(&static.Config{
		Health:   nil,
		HealthOn: false,
	})
)

//Register 注册service_http驱动工厂
func Register(register eosc.IExtenderDriverRegister) {
	register.RegisterExtenderDriver(DriverName, NewFactory())
}

type factory struct {
}

//NewFactory 创建service_http驱动工厂
func NewFactory() eosc.IExtenderDriverFactory {
	round_robin.Register()
	return &factory{}
}

//Create 创建service_http驱动
func (f *factory) Create(profession string, name string, label string, desc string, params map[string]interface{}) (eosc.IExtenderDriver, error) {

	return &driver{
		profession: profession,

		label:      label,
		desc:       desc,
		driver:     name,
		configType: reflect.TypeOf((*Config)(nil)),
	}, nil
}

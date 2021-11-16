package service_http

import (
	"reflect"

	"github.com/eolinker/eosc/common/bean"
	"github.com/eolinker/goku/plugin"

	"github.com/eolinker/goku/drivers/discovery/static"

	"github.com/eolinker/eosc"
)

var DriverName = "service_http"
var (
	defaultDiscovery = static.CreateAnonymous(&static.Config{
		Scheme:   "http",
		Health:   nil,
		HealthOn: false,
	})
	pluginManger plugin.IPluginManager
)

//Register 注册service_http驱动工厂
func Register(register eosc.IExtenderDriverRegister) {
	register.RegisterExtenderDriver(DriverName, NewFactory())
}

type factory struct {
}

//NewFactory 创建service_http驱动工厂
func NewFactory() eosc.IExtenderDriverFactory {
	return &factory{}
}

//Create 创建service_http驱动
func (f *factory) Create(profession string, name string, label string, desc string, params map[string]interface{}) (eosc.IExtenderDriver, error) {
	bean.Autowired(&pluginManger)
	return &driver{
		profession: profession,

		label:      label,
		desc:       desc,
		driver:     name,
		configType: reflect.TypeOf((*Config)(nil)),
	}, nil
}

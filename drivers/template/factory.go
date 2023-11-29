package template

import (
	"github.com/eolinker/eosc"
	"github.com/eolinker/eosc/common/bean"

	"github.com/eolinker/apinto/drivers"
	"github.com/eolinker/apinto/plugin"
)

var DriverName = "plugin_template"
var (
	pluginManger plugin.IPluginManager
)

//Register 注册service_http驱动工厂
func Register(register eosc.IExtenderDriverRegister) {
	bean.Autowired(&pluginManger)
	register.RegisterExtenderDriver(DriverName, NewFactory())
}

//NewFactory 创建service_http驱动工厂
func NewFactory() eosc.IExtenderDriverFactory {
	return drivers.NewFactory[Config](Create)
}

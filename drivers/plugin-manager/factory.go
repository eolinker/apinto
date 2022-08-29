package plugin_manager

import (
	"github.com/eolinker/apinto/plugin"
	"github.com/eolinker/eosc"
	"github.com/eolinker/eosc/common/bean"
	"github.com/eolinker/eosc/setting"
)

var (
	singleton *PluginManager
)

func init() {
	singleton = NewPluginManager()
	var i plugin.IPluginManager = singleton
	bean.Injection(&i)
}

func Register(register eosc.IExtenderDriverRegister) {
	setting.RegisterSetting("plugin", singleton)
}

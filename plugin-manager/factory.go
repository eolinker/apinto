package plugin_manager

import (
	"github.com/eolinker/apinto/plugin"
	"github.com/eolinker/eosc"
	"github.com/eolinker/eosc/common/bean"
	"github.com/eolinker/eosc/setting"
	"github.com/eolinker/eosc/utils/schema"
	"reflect"
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

func genRender() interface{} {
	render, err := schema.Generate(reflect.TypeOf((*PluginWorkerConfig)(nil)), nil)
	if err != nil {
		return nil
	}
	return render
}

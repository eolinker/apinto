package plugin_manager

import (
	"reflect"
	"sync"

	"github.com/eolinker/eosc/common/bean"
	"github.com/eolinker/goku/plugin"

	"github.com/eolinker/eosc"
)

var (
	singleton *PluginManager
	once      sync.Once
)

type PluginFactory struct {
}

func (p *PluginFactory) Check(v interface{}, workers map[eosc.RequireId]interface{}) error {
	return nil
}

func (p *PluginManager) ConfigType() reflect.Type {
	return reflect.TypeOf(new(PluginWorkerConfig))
}

func (p *PluginManager) Create(id, name string, v interface{}, workers map[eosc.RequireId]interface{}) (eosc.IWorker, error) {
	return p, nil
}

func (p *PluginFactory) Create(profession string, name string, label string, desc string, params map[string]interface{}) (eosc.IExtenderDriver, error) {

	once.Do(func() {
		singleton = NewPluginManager(profession, name)
		var i plugin.IPluginManager = singleton
		bean.Injection(&i)
	})
	return singleton, nil
}

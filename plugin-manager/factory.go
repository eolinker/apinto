package plugin_manager

import (
	"github.com/eolinker/apinto/plugin"
	"github.com/eolinker/eosc"
	"github.com/eolinker/eosc/common/bean"
	"github.com/eolinker/eosc/utils/schema"
	"reflect"
	"sync"
)

var (
	singleton *PluginManager
	once      sync.Once
)

func Register(register eosc.IExtenderDriverRegister) {
	register.RegisterExtenderDriver("plugin", NewPluginFactory())
}

type PluginFactory struct {
}

func NewPluginFactory() *PluginFactory {
	return &PluginFactory{}
}

func (f *PluginFactory) Check(v interface{}, workers map[eosc.RequireId]interface{}) error {
	return nil
}
func (f *PluginFactory) Render() interface{} {
	render, err := schema.Generate(reflect.TypeOf((*PluginWorkerConfig)(nil)), nil)
	if err != nil {
		return nil
	}
	return render
}

func (p *PluginManager) Create(id, name string, v interface{}, workers map[eosc.RequireId]eosc.IWorker) (eosc.IWorker, error) {
	p.Reset(v, workers)
	return p, nil
}

func (f *PluginFactory) Create(profession string, name string, label string, desc string, params map[string]interface{}) (eosc.IExtenderDriver, error) {

	once.Do(func() {
		singleton = NewPluginManager(profession, name)
		var i plugin.IPluginManager = singleton
		bean.Injection(&i)
	})
	return singleton, nil
}

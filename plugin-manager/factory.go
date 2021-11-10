package plugin_manager

import (
	"reflect"

	"github.com/eolinker/eosc"
	http_service "github.com/eolinker/eosc/http-service"
)

type IPluginFactory interface {
	Create(cfg interface{}) (http_service.IFilter, error)
}

type PluginFactory struct {
}

func (p *PluginManager) ConfigType() reflect.Type {
	return reflect.TypeOf(new(PluginWorkerConfig))
}

func (p *PluginManager) Create(id, name string, v interface{}, workers map[eosc.RequireId]interface{}) (eosc.IWorker, error) {
	return p, nil
}

func (p *PluginFactory) Create(profession string, name string, label string, desc string, params map[string]string) (eosc.IExtenderDriver, error) {

	pm := NewPluginManager(profession, name)

	return pm, nil
}

package setting

import (
	"errors"
	"reflect"

	plugin_manager "github.com/eolinker/goku/plugin-manager"

	"github.com/eolinker/eosc"
)

var (
	pluginWorker = "plugin"
	names        = []string{
		pluginWorker,
	}
)

//Driver 实现github.com/eolinker/eosc.eosc.IProfessionDriver接口
type Driver struct {
	configType reflect.Type
}

//NewDriver
func NewDriver(profession, name, label, desc string, params map[string]string) *Driver {
	return &Driver{
		configType: nil,
	}
}

//Create
func (h *Driver) Create(id, name string, v interface{}, workers map[eosc.RequireId]interface{}) (eosc.IWorker, error) {
	has := false
	for _, n := range names {
		if n == name {
			has = true
			break
		}
	}
	if !has {
		return nil, errors.New("create setting worker error: invalid id")
	}
	worker := &Worker{
		id:   id,
		conf: v,
	}
	switch name {
	case pluginWorker:
		{
			worker.setManager(plugin_manager.DefaultManager())
		}
	}

	return nil, nil
}

//ConfigType
func (h *Driver) ConfigType() reflect.Type {
	return h.configType
}

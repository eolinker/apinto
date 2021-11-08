package plugin_manager

import (
	"github.com/eolinker/eosc"
	"github.com/eolinker/eosc/http"
	"github.com/eolinker/goku/filter"
)

const (
	StatusDisable = "disable"
	StatusEnable  = "enable"
	StatusGlobal  = "global"
)

type IPluginManager interface {
	Create(id string, conf map[string]interface{}) http.IChain
}

type PluginManager struct {
	Factories eosc.IUntyped
}

func (p *PluginManager) Create(id string, conf map[string]interface{}) http.IChain {
	filters := make([]http.IFilter, 0, len(conf))
	return filter.CreateChain(filters)
}

type PluginObj struct {
	Name   string                 `json:"name"`
	ID     string                 `json:"id"`
	Type   string                 `json:"type"`
	Status string                 `json:"status"`
	Config map[string]interface{} `json:"config"`
}

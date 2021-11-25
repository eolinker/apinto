package plugin_manager

import (
	"fmt"

	"github.com/eolinker/eosc"
	"github.com/eolinker/goku/filter"
	"github.com/eolinker/goku/plugin"
)

type PluginObj struct {
	filter.IChainHandler
	id         string
	filterType string
	conf       map[string]*plugin.Config
	manager    eosc.IUntyped
}

func NewPluginObj(handler filter.IChainHandler, id string, filterType string, conf map[string]*plugin.Config, manager eosc.IUntyped) *PluginObj {
	obj := &PluginObj{IChainHandler: handler, id: id, filterType: filterType, conf: conf, manager: manager}

	manager.Set(fmt.Sprintf("%s:%s", id, filterType), obj)

	return obj
}

func (p *PluginObj) Destroy() {
	manager := p.manager
	if manager != nil {
		p.manager = nil
		manager.Del(fmt.Sprintf("%s:%s", p.id, p.filterType))
	}
	handler := p.IChainHandler
	if handler != nil {
		handler.Destroy()
	}
}

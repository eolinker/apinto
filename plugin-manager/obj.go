package plugin_manager

import (
	"github.com/eolinker/apinto/filter"
	"github.com/eolinker/apinto/plugin"
	"github.com/eolinker/eosc"
)

type PluginObj struct {
	filter.IChainHandler
	id      string
	conf    map[string]*plugin.Config
	manager eosc.IUntyped
}

func NewPluginObj(handler filter.IChainHandler, id string, conf map[string]*plugin.Config, manager eosc.IUntyped) *PluginObj {
	obj := &PluginObj{IChainHandler: handler, id: id, conf: conf, manager: manager}

	manager.Set(id, obj)

	return obj
}

func (p *PluginObj) Destroy() {
	manager := p.manager
	if manager != nil {
		p.manager = nil
		manager.Del(p.id)
	}
	handler := p.IChainHandler
	if handler != nil {
		handler.Destroy()
	}
}

package plugin_manager

import (
	"github.com/eolinker/goku/filter"
	"github.com/eolinker/goku/plugin"
)

type PluginObj struct {
	filter.IChainHandler
	id         string
	filterType string
	conf       map[string]*plugin.Config

	manager *PluginManager
}

func (p *PluginObj) Destroy() {
	p.manager.RemoveObj(p.id)
}

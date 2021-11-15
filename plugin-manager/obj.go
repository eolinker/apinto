package plugin_manager

import "github.com/eolinker/goku/filter"

type PluginObj struct {
	filter.IChainHandler
	id   string
	t    string
	conf map[string]*OrdinaryPlugin

	manager *PluginManager
}

func (p *PluginObj) Destroy() {
	p.manager.RemoveObj(p.id)
}

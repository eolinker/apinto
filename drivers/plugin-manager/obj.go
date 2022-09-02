package plugin_manager

import (
	"github.com/eolinker/apinto/plugin"
	eoscContext "github.com/eolinker/eosc/eocontext"
)

type PluginObj struct {
	eoscContext.Filters
	id   string
	conf map[string]*plugin.Config
}

func NewPluginObj(filters eoscContext.Filters, id string, conf map[string]*plugin.Config) *PluginObj {
	obj := &PluginObj{Filters: filters, id: id, conf: conf}

	return obj
}

func (p *PluginObj) Destroy() {

	handler := p.Filters
	if handler != nil {
		handler.Destroy()
	}
}

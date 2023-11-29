package plugin_manager

import (
	eoscContext "github.com/eolinker/eosc/eocontext"

	"github.com/eolinker/apinto/plugin"
)

type PluginObj struct {
	fs   eoscContext.Filters
	id   string
	conf map[string]*plugin.Config
}

func NewPluginObj(filters eoscContext.Filters, id string, conf map[string]*plugin.Config) *PluginObj {
	obj := &PluginObj{fs: filters, id: id, conf: conf}

	return obj
}

func (p *PluginObj) Chain(ctx eoscContext.EoContext, append ...eoscContext.IFilter) error {
	return eoscContext.DoChain(ctx, p.fs, append...)
}
func (p *PluginObj) Destroy() {

	handler := p.fs
	if handler != nil {
		handler.Destroy()
	}
}

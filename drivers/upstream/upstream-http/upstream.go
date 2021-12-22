package upstream_http

import (
	"time"

	"github.com/eolinker/goku/upstream"

	"github.com/eolinker/eosc"

	"github.com/eolinker/goku/discovery"
	"github.com/eolinker/goku/plugin"
	"github.com/eolinker/goku/upstream/balance"
)

type Upstream struct {
	scheme  string
	app     discovery.IApp
	handler balance.IBalanceHandler

	handlers eosc.IUntyped

	pluginConf map[string]*plugin.Config
}

func (up *Upstream) Create(id string, configs map[string]*plugin.Config, retry int, timeout time.Duration) (upstream.IUpstreamHandler, error) {
	return up.create(id, configs, retry, timeout), nil
}
func (up *Upstream) create(id string, configs map[string]*plugin.Config, retry int, timeout time.Duration) *UpstreamHandler {
	nh := NewUpstreamHandler(id, up, retry, timeout, configs)
	up.handlers.Set(id, nh)
	return nh
}

func (up *Upstream) Merge(configs map[string]*plugin.Config) map[string]*plugin.Config {
	return plugin.MergeConfig(configs, up.pluginConf)
}
func NewUpstream(scheme string, app discovery.IApp, handler balance.IBalanceHandler, pluginConf map[string]*plugin.Config) *Upstream {
	return &Upstream{scheme: scheme, app: app, handler: handler, handlers: eosc.NewUntyped(), pluginConf: pluginConf}
}

//Reset reset
func (up *Upstream) Reset(scheme string, app discovery.IApp, handler balance.IBalanceHandler, pluginConf map[string]*plugin.Config) {
	up.scheme = scheme
	up.app = app
	up.handler = handler
	up.pluginConf = pluginConf

	for _, h := range up.handlers.List() {
		hd := h.(*UpstreamHandler)
		hd.reset()
	}
}

func (up *Upstream) destroy() {
	handlers := up.handlers.List()
	up.handlers = eosc.NewUntyped()
	for _, h := range handlers {
		hd := h.(*UpstreamHandler)
		hd.orgFilter.Destroy()
	}

}

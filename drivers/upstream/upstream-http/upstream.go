package upstream_http

import (
	"time"

	"github.com/eolinker/apinto/upstream"

	"github.com/eolinker/eosc"

	"github.com/eolinker/apinto/discovery"
	"github.com/eolinker/apinto/upstream/balance"
)

type Upstream struct {
	scheme  string
	app     discovery.IApp
	handler balance.IBalanceHandler

	handlers eosc.IUntyped
}

func (up *Upstream) Create(id string, retry int, timeout time.Duration) (upstream.IUpstreamHandler, error) {
	return up.create(id, retry, timeout), nil
}
func (up *Upstream) create(id string, retry int, timeout time.Duration) *UpstreamHandler {
	nh := NewUpstreamHandler(id, up, retry, timeout)
	up.handlers.Set(id, nh)
	return nh
}

func NewUpstream(scheme string, app discovery.IApp, handler balance.IBalanceHandler) *Upstream {
	return &Upstream{scheme: scheme, app: app, handler: handler, handlers: eosc.NewUntyped()}
}

//Reset reset
func (up *Upstream) Reset(scheme string, app discovery.IApp, handler balance.IBalanceHandler) {
	up.scheme = scheme
	up.app = app
	up.handler = handler
}

func (up *Upstream) destroy() {
	handlers := up.handlers.List()
	up.handlers = eosc.NewUntyped()
	for _, h := range handlers {
		hd := h.(*UpstreamHandler)
		hd.Destroy()
	}

}

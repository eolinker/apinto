package service_http

import (
	http_service "github.com/eolinker/eosc/http-service"
	"github.com/eolinker/goku/plugin"
	"github.com/eolinker/goku/upstream"
)

type ServiceHandler struct {
	service    *Service
	id         string
	config     map[string]*plugin.Config
	pluginExec upstream.IUpstreamHandler

	proxyMethod string
}

func (s *ServiceHandler) DoChain(ctx http_service.IHttpContext) error {
	if s.proxyMethod != "" {
		ctx.Proxy().SetMethod(s.proxyMethod)
	}
	if s.pluginExec != nil {
		s.pluginExec.DoChain(ctx)
	}
	return nil
}

func (s *ServiceHandler) Destroy() {
	if s.pluginExec != nil {
		s.pluginExec.Destroy()
		s.pluginExec = nil
	}
}

func (s *ServiceHandler) rebuild(upstream upstream.IUpstream) {

	ps, err := upstream.Create(s.id, s.service.mergePluginConfig(s.config), s.service.retry, s.service.timeout)
	if err != nil {
		return
	}
	s.pluginExec = ps
}

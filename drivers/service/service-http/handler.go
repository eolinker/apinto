package service_http

import (
	http_service "github.com/eolinker/eosc/http-service"
	"github.com/eolinker/goku/filter"
	"github.com/eolinker/goku/plugin"
	"github.com/eolinker/goku/upstream"
)

type ServiceHandler struct {
	service         *Service
	id              string
	config          map[string]*plugin.Config
	pluginExec      http_service.IChain
	upstreamHandler upstream.IUpstreamHandler
}

func (s *ServiceHandler) DoFilter(ctx http_service.IHttpContext, next http_service.IChain) (err error) {
	if s.upstreamHandler != nil {
		err = s.upstreamHandler.DoChain(ctx)
	}
	if err == nil && next != nil {
		err = next.DoChain(ctx)
	}
	return
}

func (s *ServiceHandler) DoChain(ctx http_service.IHttpContext) error {
	if s.service.proxyMethod != "" {
		ctx.Proxy().SetMethod(s.service.proxyMethod)
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
	serviceFilter := pluginManger.CreateService(s.id, s.config)
	s.pluginExec = serviceFilter.Append(filter.ToFilter([]http_service.IFilter{s}))

	ps, err := upstream.Create(s.id, s.service.mergePluginConfig(s.config), s.service.retry, s.service.timeout)
	if err != nil {
		return
	}
	s.upstreamHandler = ps

}

package service_http

import (
	http_service "github.com/eolinker/eosc/http-service"
	"github.com/eolinker/eosc/log"
	"github.com/eolinker/goku/filter"
	"github.com/eolinker/goku/plugin"
	"github.com/eolinker/goku/upstream"
)

type ServiceHandler struct {
	service         *Service
	id              string
	config          map[string]*plugin.Config
	pluginExec      http_service.IChain
	pluginOrg       plugin.IPlugin
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
	service := s.service
	if service == nil {
		return nil
	}
	if service.proxyMethod != "" {
		ctx.Proxy().SetMethod(service.proxyMethod)
	}
	exec := s.pluginExec
	if exec != nil {
		return exec.DoChain(ctx)
	}
	return nil
}

func (s *ServiceHandler) Destroy() {
	plg := s.pluginOrg
	if plg != nil {
		s.pluginOrg = nil
		plg.Destroy()
	}
	ser := s.service
	if ser != nil {
		s.service = nil
		ser.handlers.Del(s.id)
	}
}

func (s *ServiceHandler) rebuild(upstream upstream.IUpstream) {
	s.pluginOrg = pluginManger.CreateService(s.id, s.service.mergePluginConfig(s.config))
	s.pluginExec = s.pluginOrg.Append(filter.ToFilter([]http_service.IFilter{s}))

	ps, err := upstream.Create(s.id, s.service.mergePluginConfig(s.config), s.service.retry, s.service.timeout)
	if err != nil {
		log.Error("rebuild error: ", err)
		return
	}
	s.upstreamHandler = ps

}

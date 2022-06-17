package service_http

import (
	"github.com/eolinker/apinto/filter"
	"github.com/eolinker/apinto/plugin"
	"github.com/eolinker/apinto/upstream"
	http_service "github.com/eolinker/eosc/http-service"
	"github.com/eolinker/eosc/log"
)

type ServiceHandler struct {
	service            *Service
	id                 string
	routerPluginConfig map[string]*plugin.Config
	pluginExec         http_service.IChain
	pluginOrg          plugin.IPlugin
	upstreamHandler    upstream.IUpstreamHandler
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

func (s *ServiceHandler) rebuild() {
	config := s.service.Merge(s.routerPluginConfig)

	s.pluginOrg = pluginManger.CreateRequest(s.id, config)
	s.pluginExec = s.pluginOrg.Append(filter.ToFilter([]http_service.IFilter{s}))

	ps, err := s.service.upstream.Create(s.id, s.service.retry, s.service.timeout)
	if err != nil {
		log.Error("rebuild error: ", err)
		return
	}
	s.upstreamHandler = ps
}

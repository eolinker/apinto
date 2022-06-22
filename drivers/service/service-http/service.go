package service_http

import (
	"github.com/eolinker/apinto/discovery"
	"github.com/eolinker/apinto/upstream/balance"
	"time"

	"github.com/eolinker/eosc/log"

	"github.com/eolinker/apinto/plugin"
	"github.com/eolinker/apinto/service"
)

type Service struct {
	upstream *Upstream
	configs  map[string]*plugin.Config
	handlers *Handlers
	retry    int
	timeout  time.Duration
}

func (s *Service) reset(scheme string, app discovery.IApp, handler balance.IBalanceHandler, config map[string]*plugin.Config) {
	s.configs = config
	if s.upstream == nil {
		s.upstream = NewUpstream(scheme, app, handler)
	} else {
		s.upstream.Reset(scheme, app, handler)
	}

	log.Debug("reset upstream handler...handler size is ", len(s.handlers.List()))
	for _, h := range s.handlers.List() {
		h.rebuild()
	}
}
func (s *Service) Merge(config map[string]*plugin.Config) map[string]*plugin.Config {
	configs := plugin.MergeConfig(config, s.configs)

	return configs
}
func (s *Service) Create(id string, configs map[string]*plugin.Config) service.IService {
	h := s.newHandler(id, configs)
	h.rebuild()
	s.handlers.Set(id, h)
	return h
}

func (s *Service) newHandler(id string, config map[string]*plugin.Config) *ServiceHandler {
	return &ServiceHandler{
		service:            s,
		id:                 id,
		routerPluginConfig: config,
	}
}

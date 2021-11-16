package service_http

import (
	"time"

	"github.com/eolinker/goku/plugin"
	"github.com/eolinker/goku/service"
	"github.com/eolinker/goku/upstream"
)

type Service struct {
	upstream upstream.IUpstream
	configs  map[string]*plugin.Config
	handlers *Handlers
	retry    int
	timeout  time.Duration

	scheme      string
	proxyMethod string
}

func (s *Service) reset(upstream upstream.IUpstream, config map[string]*plugin.Config) {
	s.configs = config
	s.upstream = upstream

	for _, h := range s.handlers.List() {
		h.rebuild(upstream)
	}
}
func (s *Service) mergePluginConfig(config map[string]*plugin.Config) map[string]*plugin.Config {
	return plugin.MergeConfig(config, s.configs)
}
func (s *Service) Create(id string, configs map[string]*plugin.Config) service.IService {
	h := s.newHandler(id, configs)
	h.rebuild(s.upstream)
	return nil
}

func (s *Service) newHandler(id string, config map[string]*plugin.Config) *ServiceHandler {
	return &ServiceHandler{
		service: s,
		id:      id,
		config:  config,
	}
}

package service_http

import (
	"time"

	"github.com/eolinker/eosc/log"

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
	log.Debug("reset upstream handler...handler size is ", len(s.handlers.List()))
	for _, h := range s.handlers.List() {
		h.rebuild()
	}
}
func (s *Service) Merge(config map[string]*plugin.Config) map[string]*plugin.Config {
	configs := plugin.MergeConfig(config, s.configs)
	if mg, ok := s.upstream.(plugin.IPluginConfigMerge); ok {
		config = mg.Merge(config)
	}
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

package service_http

import (
	"github.com/eolinker/goku/plugin"
	"github.com/eolinker/goku/service"
	"github.com/eolinker/goku/upstream"
)

type Service struct {
	upstream upstream.IUpstream
	configs  map[string]*plugin.Config
	handlers Handlers
}

func NewService() *Service {
	return &Service{}
}
func (s *Service) reset(upstream upstream.IUpstream, config Config) {

}
func (s *Service) Create(id string, configs map[string]*plugin.Config) service.IService {

}

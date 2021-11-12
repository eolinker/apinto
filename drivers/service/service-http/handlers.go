package service_http

import (
	"github.com/eolinker/eosc"
	"github.com/eolinker/goku/plugin"
)

type IHandlers interface {
	Get(id string) (*ServiceHandler, bool)
	Set(id string, handler *ServiceHandler)
	Del(id string) (*ServiceHandler, bool)
}
type Handlers struct {
	data eosc.IUntyped
}

func NewHandlers() *Handlers {
	return &Handlers{
		data: eosc.NewUntyped(),
	}
}

func (s *serviceWorker) Create(id string, configs map[string]*plugin.Config) plugin.IPlugin {

}

package service_http

import (
	"github.com/eolinker/eosc"
)

type IHandlers interface {
	Set(id string, handler *ServiceHandler)
	Del(id string) (*ServiceHandler, bool)
	List() []*ServiceHandler
}
type Handlers struct {
	data eosc.IUntyped
}

func (h *Handlers) List() []*ServiceHandler {
	list := h.data.List()
	rs := make([]*ServiceHandler, len(list))
	for i, v := range list {
		rs[i] = v.(*ServiceHandler)
	}
	return rs
}

func (h *Handlers) Set(id string, handler *ServiceHandler) {
	h.data.Set(id, handler)
}

func (h *Handlers) Del(id string) (*ServiceHandler, bool) {
	v, has := h.data.Del(id)
	if has {
		return v.(*ServiceHandler), true
	}
	return nil, false
}

func NewHandlers() *Handlers {
	return &Handlers{
		data: eosc.NewUntyped(),
	}
}

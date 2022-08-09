package http_router

import (
	"github.com/eolinker/eosc"
	eoscContext "github.com/eolinker/eosc/eocontext"
)

type HttpRouter struct {
	id   string
	name string

	config *Config

	filter *eoscContext.IChain
}

func (h *HttpRouter) Id() string {
	return h.id
}

func (h *HttpRouter) Start() error {
	return nil
}

func (h *HttpRouter) Reset(conf interface{}, workers map[eosc.RequireId]eosc.IWorker) error {
	err := h.reset(conf, workers)
	if err != nil {
		return err
	}
	return nil
}
func (h *HttpRouter) reset(conf interface{}, workers map[eosc.RequireId]eosc.IWorker) error {
	return nil
}
func (h *HttpRouter) Stop() error {
	return nil
}

func (h *HttpRouter) CheckSkill(skill string) bool {
	return false
}

package http_router

import (
	service "github.com/eolinker/apinto/v2"
	"github.com/eolinker/eosc"
	"time"
)

type HttpRouter struct {
	id   string
	name string

	handler *Handler
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
	cfg, ok := conf.(*Config)
	if !ok {
		return eosc.ErrorConfigFieldUnknown
	}
	serviceWorker, has := workers[cfg.Service]
	if !has || !serviceWorker.CheckSkill(service.ServiceSkill) {
		return eosc.ErrorNotGetSillForRequire
	}

	templateWorker, has := workers[cfg.Template]
	if !has || !templateWorker.CheckSkill(service.TemplateSkill) {
		return eosc.ErrorNotGetSillForRequire
	}
	template := templateWorker.(service.ITemplate)
	plugins := template.Create(h.id, cfg.Plugins)
	serviceHandler := serviceWorker.(service.IService)

	h.handler = &Handler{
		completeHandler: HttpComplete{
			retry:   cfg.Retry,
			timeOut: time.Duration(cfg.TimeOut) * time.Millisecond,
		},
		finisher: Finisher{},
		service:  serviceHandler,
		filters:  plugins,
	}
	return nil
}
func (h *HttpRouter) Stop() error {
	return nil
}

func (h *HttpRouter) CheckSkill(skill string) bool {
	return false
}

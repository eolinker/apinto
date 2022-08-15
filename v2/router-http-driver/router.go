package http_router

import (
	"github.com/eolinker/apinto/plugin"
	service "github.com/eolinker/apinto/v2"
	"github.com/eolinker/apinto/v2/router"
	router_http_manager "github.com/eolinker/apinto/v2/router-http-manager"
	"github.com/eolinker/eosc"
	"github.com/eolinker/eosc/eocontext"
	"time"
)

type HttpRouter struct {
	id   string
	name string

	handler *Handler

	pluginManager plugin.IPluginManager
	routerManager router_http_manager.IManger
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

	if cfg.Plugins == nil {
		cfg.Plugins = map[string]*plugin.Config{}
	}
	var plugins eocontext.IChain
	if cfg.Template != "" {
		templateWorker, has := workers[cfg.Template]
		if !has || !templateWorker.CheckSkill(service.TemplateSkill) {
			return eosc.ErrorNotGetSillForRequire
		}
		template := templateWorker.(service.ITemplate)
		plugins = template.Create(h.id, cfg.Plugins)
	} else {
		plugins = h.pluginManager.CreateRequest(h.id, cfg.Plugins)
	}

	serviceHandler := serviceWorker.(service.IService)

	handler := &Handler{
		completeHandler: HttpComplete{
			retry:   cfg.Retry,
			timeOut: time.Duration(cfg.TimeOut) * time.Millisecond,
		},
		finisher: Finisher{},
		service:  serviceHandler,
		filters:  plugins,
	}
	appendRule := make([]router.AppendRule, 0, len(cfg.Rules))
	for _, r := range cfg.Rules {
		appendRule = append(appendRule, router.AppendRule{
			Type:    r.Type,
			Name:    r.Name,
			Pattern: r.Value,
		})
	}
	err := h.routerManager.Set(h.id, cfg.Listen, cfg.Host, cfg.Method, cfg.Path, appendRule, handler)
	if err != nil {
		return err
	}
	h.handler = handler
	return nil
}
func (h *HttpRouter) Stop() error {
	return nil
}

func (h *HttpRouter) CheckSkill(skill string) bool {
	return false
}

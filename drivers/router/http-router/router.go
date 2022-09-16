package http_router

import (
	http_complete "github.com/eolinker/apinto/drivers/router/http-router/http-complete"
	"github.com/eolinker/apinto/drivers/router/http-router/manager"
	"github.com/eolinker/apinto/plugin"
	http_router "github.com/eolinker/apinto/router/http-router"
	"github.com/eolinker/apinto/service"
	"github.com/eolinker/apinto/template"
	"strings"
	"time"

	"github.com/eolinker/eosc"
	"github.com/eolinker/eosc/eocontext"
)

type HttpRouter struct {
	id            string
	name          string
	routerManager manager.IManger
	pluginManager plugin.IPluginManager
	handler       *Handler
}

func (h *HttpRouter) Destroy() error {

	h.routerManager.Delete(h.id)
	return nil
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
	handler := &Handler{
		routerName:      h.name,
		serviceName:     strings.TrimSuffix(string(cfg.Service), "@service"),
		completeHandler: http_complete.NewHttpComplete(cfg.Retry, time.Duration(cfg.TimeOut)*time.Millisecond),
		finisher:        Finisher{},
		service:         nil,
		filters:         nil,
		disable:         cfg.Disable,
	}
	if !cfg.Disable {

		serviceWorker, has := workers[cfg.Service]
		if !has || !serviceWorker.CheckSkill(service.ServiceSkill) {
			return eosc.ErrorNotGetSillForRequire
		}

		if cfg.Plugins == nil {
			cfg.Plugins = map[string]*plugin.Config{}
		}
		var plugins eocontext.IChainPro
		if cfg.Template != "" {
			templateWorker, has := workers[cfg.Template]
			if !has || !templateWorker.CheckSkill(template.TemplateSkill) {
				return eosc.ErrorNotGetSillForRequire
			}
			tp := templateWorker.(template.ITemplate)
			plugins = tp.Create(h.id, cfg.Plugins)
		} else {
			plugins = h.pluginManager.CreateRequest(h.id, cfg.Plugins)
		}

		serviceHandler := serviceWorker.(service.IService)

		handler.service = serviceHandler
		handler.filters = plugins
	}

	appendRule := make([]http_router.AppendRule, 0, len(cfg.Rules))
	for _, r := range cfg.Rules {
		appendRule = append(appendRule, http_router.AppendRule{
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
	h.Destroy()
	return nil
}

func (h *HttpRouter) CheckSkill(skill string) bool {
	return false
}

package http_router

import (
	"net/http"
	"strings"
	"time"

	"github.com/eolinker/apinto/drivers/router/http-router/websocket"

	"github.com/eolinker/apinto/service"

	"github.com/eolinker/eosc/eocontext"

	"github.com/eolinker/apinto/drivers"
	http_complete "github.com/eolinker/apinto/drivers/router/http-router/http-complete"
	"github.com/eolinker/apinto/drivers/router/http-router/manager"
	"github.com/eolinker/apinto/plugin"
	http_router "github.com/eolinker/apinto/router"
	"github.com/eolinker/apinto/template"
	"github.com/eolinker/eosc"
)

type HttpRouter struct {
	id            string
	name          string
	routerManager manager.IManger
	pluginManager plugin.IPluginManager
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
	cfg, err := drivers.Assert[Config](conf)
	if err != nil {
		return err
	}
	return h.reset(cfg, workers)

}
func (h *HttpRouter) reset(cfg *Config, workers map[eosc.RequireId]eosc.IWorker) error {

	methods := cfg.Method

	handler := &httpHandler{
		routerName:  h.name,
		routerId:    h.id,
		serviceName: strings.TrimSuffix(string(cfg.Service), "@service"),
		finisher:    defaultFinisher,
		disable:     cfg.Disable,
		websocket:   cfg.Websocket,
	}

	if !cfg.Disable {

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
		handler.filters = plugins

		if cfg.Service == "" {
			// 当service未指定，使用默认返回
			handler.completeHandler = http_complete.NewNoServiceCompleteHandler(cfg.Status, cfg.Header, cfg.Body)
		} else {
			serviceWorker, has := workers[cfg.Service]
			if !has || !serviceWorker.CheckSkill(service.ServiceSkill) {
				return eosc.ErrorNotGetSillForRequire
			}
			serviceHandler := serviceWorker.(service.IService)
			handler.service = serviceHandler
			if cfg.Websocket {
				handler.completeHandler = websocket.NewComplete(cfg.Retry, time.Duration(cfg.TimeOut)*time.Millisecond)
				methods = []string{http.MethodGet}
			} else {
				handler.completeHandler = http_complete.NewHttpComplete(cfg.Retry, time.Duration(cfg.TimeOut)*time.Millisecond)
			}
		}
	}

	appendRule := make([]http_router.AppendRule, 0, len(cfg.Rules))
	for _, r := range cfg.Rules {
		appendRule = append(appendRule, http_router.AppendRule{
			Type:    r.Type,
			Name:    r.Name,
			Pattern: r.Value,
		})
	}
	err := h.routerManager.Set(h.id, cfg.Listen, cfg.Host, methods, cfg.Path, appendRule, handler)
	if err != nil {
		return err
	}
	return nil
}
func (h *HttpRouter) Stop() error {
	h.Destroy()
	return nil
}

func (h *HttpRouter) CheckSkill(skill string) bool {
	return false
}

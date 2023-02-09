package grpc_router

import (
	"strings"
	"time"

	"github.com/eolinker/apinto/drivers"
	"github.com/eolinker/apinto/router"

	"github.com/eolinker/apinto/drivers/router/grpc-router/manager"
	"github.com/eolinker/apinto/plugin"
	"github.com/eolinker/apinto/service"
	"github.com/eolinker/apinto/template"

	"github.com/eolinker/eosc"
	"github.com/eolinker/eosc/eocontext"
)

type GrpcRouter struct {
	id            string
	name          string
	routerManager manager.IManger
	pluginManager plugin.IPluginManager
}

func (h *GrpcRouter) Destroy() error {

	h.routerManager.Delete(h.id)
	return nil
}

func (h *GrpcRouter) Id() string {
	return h.id
}

func (h *GrpcRouter) Start() error {
	return nil
}

func (h *GrpcRouter) Reset(conf interface{}, workers map[eosc.RequireId]eosc.IWorker) error {
	cfg, err := drivers.Assert[Config](conf)
	if err != nil {
		return err
	}
	return h.reset(cfg, workers)

}
func (h *GrpcRouter) reset(cfg *Config, workers map[eosc.RequireId]eosc.IWorker) error {

	handler := &grpcRouter{
		routerName:      h.name,
		routerId:        h.id,
		serviceName:     strings.TrimSuffix(string(cfg.Service), "@service"),
		completeHandler: NewComplete(cfg.Retry, time.Duration(cfg.TimeOut)*time.Millisecond),
		finisher:        defaultFinisher,
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

	appendRule := make([]router.AppendRule, 0, len(cfg.Rules))
	for _, r := range cfg.Rules {
		appendRule = append(appendRule, router.AppendRule{
			Type:    r.Type,
			Name:    r.Name,
			Pattern: r.Value,
		})
	}
	err := h.routerManager.Set(h.id, cfg.Listen, cfg.Host, cfg.ServiceName, cfg.MethodName, appendRule, handler)
	if err != nil {
		return err
	}
	return nil
}
func (h *GrpcRouter) Stop() error {
	h.Destroy()
	return nil
}

func (h *GrpcRouter) CheckSkill(skill string) bool {
	return false
}

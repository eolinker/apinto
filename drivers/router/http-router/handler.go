package http_router

import (
	"fmt"
	http_service "github.com/eolinker/eosc/http-service"
	service "github.com/eolinker/eosc/http-service"
	"github.com/eolinker/apinto/plugin"
	router_http "github.com/eolinker/apinto/router/router-http"
	service2 "github.com/eolinker/apinto/service"
)

type RouterHandler struct {
	routerConfig  *router_http.Config
	routerFilters http_service.IChain
	serviceFilter service2.IService
}

func (r *RouterHandler) DoFilter(ctx service.IHttpContext, next service.IChain) (err error) {
	return r.serviceFilter.DoChain(ctx)
}

func (r *RouterHandler) Destroy() {
	s := r.serviceFilter
	if s != nil {
		r.serviceFilter = nil
		s.Destroy()
	}
	rh := r.routerFilters
	if rh != nil {
		r.routerFilters = nil
		rh.Destroy()
	}
}

func NewRouterHandler(routerConfig *router_http.Config, routerPlugin plugin.IPlugin, handler service2.IService) *RouterHandler {

	r := &RouterHandler{routerConfig: routerConfig, serviceFilter: handler}
	r.routerFilters = routerPlugin.Append(r)
	routerConfig.Target = r.routerFilters
	return r
}

func NewDisableHandler(routerConfig *router_http.Config) *RouterHandler {
	r := &RouterHandler{routerConfig: routerConfig, serviceFilter: &DisableHandler{}}
	routerConfig.Target = r.serviceFilter
	return r
}

type DisableHandler struct {
}

func (d *DisableHandler) DoChain(ctx service.IHttpContext) error {
	resp := ctx.Response()
	resp.SetBody([]byte("the router is disabled"))
	resp.SetStatus(416, "416")
	return fmt.Errorf("the router is disabled")
}

func (d *DisableHandler) Destroy() {
	return
}

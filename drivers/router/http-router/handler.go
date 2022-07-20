package http_router

import (
	"fmt"
	router_http "github.com/eolinker/apinto/router/router-http"
	service2 "github.com/eolinker/apinto/service"
	"github.com/eolinker/eosc/context"
	service "github.com/eolinker/eosc/context/http-context"
)

var _ service.HttpFilter = (*RouterHandler)(nil)

type RouterHandler struct {
	routerConfig  *router_http.Config
	serviceFilter service2.IService
}

func (r *RouterHandler) DoHttpFilter(ctx service.IHttpContext, next context.IChain) (err error) {
	return r.serviceFilter.DoChain(ctx)
}

func (r *RouterHandler) Destroy() {
	s := r.serviceFilter
	if s != nil {
		r.serviceFilter = nil
		s.Destroy()
	}

}

func NewRouterHandler(routerConfig *router_http.Config, handler service2.IService) *RouterHandler {

	r := &RouterHandler{routerConfig: routerConfig, serviceFilter: handler}
	routerConfig.Target = handler
	return r
}

func NewDisableHandler(routerConfig *router_http.Config) *RouterHandler {
	r := &RouterHandler{routerConfig: routerConfig, serviceFilter: &DisableHandler{}}
	routerConfig.Target = r.serviceFilter
	return r
}

type DisableHandler struct {
}

func (d *DisableHandler) DoChain(ctx context.Context) error {
	httpContext, err := service.Assert(ctx)
	if err != nil {
		return err
	}
	resp := httpContext.Response()
	resp.SetBody([]byte("the router is disabled"))
	resp.SetStatus(416, "416")
	return fmt.Errorf("the router is disabled")
}

func (d *DisableHandler) Destroy() {
	return
}

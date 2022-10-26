package http_router

import (
	"net/http"
	"strings"

	http_complete "github.com/eolinker/apinto/drivers/router/http-router/http-complete"
	"github.com/eolinker/apinto/service"
	"github.com/eolinker/eosc/eocontext"
	http_context "github.com/eolinker/eosc/eocontext/http-context"
)

var completeCaller = http_complete.NewHttpCompleteCaller()

type Handler struct {
	completeHandler *http_complete.HttpComplete

	routerName  string
	serviceName string

	finisher  Finisher
	service   service.IService
	filters   eocontext.IChainPro
	disable   bool
	websocket bool
}

func (h *Handler) ServeHTTP(ctx eocontext.EoContext) {

	httpContext, err := http_context.Assert(ctx)
	if err != nil {
		return
	}
	if h.websocket {
		if strings.Contains(httpContext.Request().Header().GetHeader("Connection"), "upgrade") && strings.Contains(httpContext.Request().Header().GetHeader("Upgrade"), "websocket") {

		}
	}

	if h.disable {
		httpContext.Response().SetStatus(http.StatusNotFound, "")
		httpContext.Response().SetBody([]byte("router disable"))
		httpContext.FastFinish()
		return
	}
	//Set Label
	ctx.SetLabel("api", h.routerName)
	ctx.SetLabel("service", h.serviceName)
	ctx.SetLabel("ip", httpContext.Request().ReadIP())
	ctx.SetFinish(&h.finisher)
	ctx.SetCompleteHandler(h.completeHandler)
	ctx.SetApp(h.service)
	ctx.SetBalance(h.service)
	ctx.SetUpstreamHostHandler(h.service)

	h.filters.Chain(ctx, completeCaller)
}

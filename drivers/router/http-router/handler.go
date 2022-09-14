package http_router

import (
	"github.com/eolinker/apinto/service"
	"github.com/eolinker/eosc/eocontext"
	http_context "github.com/eolinker/eosc/eocontext/http-context"
	"net/http"
)

type Handler struct {
	completeHandler HttpComplete
	finisher        Finisher
	service         service.IService
	filters         eocontext.IChain
	disable         bool
	labels          map[string]string
}

func (h *Handler) ServeHTTP(ctx eocontext.EoContext) {
	httpContext, err := http_context.Assert(ctx)
	if err != nil {
		return
	}
	if h.disable {
		httpContext.Response().SetStatus(http.StatusNotFound, "")
		httpContext.Response().SetBody([]byte("router disable"))
		httpContext.FastFinish()
		return
	}
	//Set Label
	ctx.SetLabel("api", h.labels["api"])
	ctx.SetLabel("service", h.labels["service"])
	ctx.SetLabel("path", httpContext.Request().URI().Path())
	ctx.SetLabel("method", httpContext.Request().Method())

	ctx.SetFinish(&h.finisher)
	ctx.SetCompleteHandler(&h.completeHandler)
	ctx.SetApp(h.service)
	ctx.SetBalance(h.service)
	h.filters.DoChain(ctx)
}

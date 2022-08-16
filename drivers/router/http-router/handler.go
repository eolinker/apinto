package http_router

import (
	"github.com/eolinker/apinto/service"
	"github.com/eolinker/eosc/eocontext"
)

type Handler struct {
	completeHandler HttpComplete
	finisher        Finisher
	service         service.IService
	filters         eocontext.IChain
}

func (h *Handler) ServeHTTP(ctx eocontext.EoContext) {
	ctx.SetFinish(&h.finisher)
	ctx.SetCompleteHandler(&h.completeHandler)
	ctx.SetApp(h.service)
	ctx.SetBalance(h.service)
	h.filters.DoChain(ctx)
}

package grpc_router

import (
	"github.com/eolinker/apinto/service"

	"github.com/eolinker/eosc/eocontext"
)

var completeCaller = NewCompleteCaller()

type grpcRouter struct {
	completeHandler eocontext.CompleteHandler

	routerName  string
	routerId    string
	serviceName string

	finisher eocontext.FinishHandler
	service  service.IService
	filters  eocontext.IChainPro
	disable  bool
}

func (h *grpcRouter) ServeHTTP(ctx eocontext.EoContext) {

}

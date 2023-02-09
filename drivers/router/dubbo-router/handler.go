package dubbo_router

import (
	dubbo_router "github.com/eolinker/apinto/router/dubbo-router"
	"github.com/eolinker/apinto/service"
	"github.com/eolinker/eosc/eocontext"
	dubbo_context "github.com/eolinker/eosc/eocontext/dubbo-context"
)

var _ dubbo_router.IDubboRouterHandler = (*dubboHandler)(nil)

type dubboHandler struct {
	completeHandler eocontext.CompleteHandler
	routerName      string
	routerId        string
	serviceName     string
	disable         bool
	service         service.IService
}

func (d *dubboHandler) DubboProxy(ctx dubbo_context.IDubboContext) {
}

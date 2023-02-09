package dubbo_router

import (
	dubbo_context "github.com/eolinker/eosc/eocontext/dubbo-context"
)

type IMatcher interface {
	Match(port int, service dubbo_context.IRequestReader) (IDubboRouterHandler, bool)
}

type IDubboRouterHandler interface {
	DubboProxy(ctx dubbo_context.IDubboContext)
}

package dubbo_router

import (
	"dubbo.apache.org/dubbo-go/v3/protocol/dubbo/impl"
)

type IMatcher interface {
	Match(port int, service impl.Service) (IDubboRouterHandler, bool)
}

type IDubboRouterHandler interface {
	DubboProxy(dubboPackage *impl.DubboPackage)
}

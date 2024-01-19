package router

import (
	eoscContext "github.com/eolinker/eosc/eocontext"
)

type IMatcher interface {
	Match(port int, request interface{}) (IRouterHandler, bool)
}

type IRouterHandler interface {
	Serve(ctx eoscContext.EoContext)
}

type IRouterPreHandler interface {
	Server(ctx eoscContext.EoContext) (isContinue bool)
}

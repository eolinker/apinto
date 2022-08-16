package http_router

import (
	eoscContext "github.com/eolinker/eosc/eocontext"
	http_service "github.com/eolinker/eosc/eocontext/http-context"
)

type IMatcher interface {
	Match(port int, request http_service.IRequestReader) (IRouterHandler, bool)
}

type IRouterHandler interface {
	ServeHTTP(ctx eoscContext.EoContext)
}

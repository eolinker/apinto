package http_context

import (
	http_service "github.com/eolinker/eosc/eocontext/http-context"
	"sync"
)

var (
	pool sync.Pool = sync.Pool{
		New: newContext,
	}
)

func newContext() interface{} {
	h := new(HttpContext)
	h.proxyRequests = make([]http_service.IRequest, 0, 5)
	return h
}

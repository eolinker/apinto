package http_context

import (
	"sync"

	http_service "github.com/eolinker/eosc/eocontext/http-context"
)

var (
	pool = sync.Pool{
		New: newContext,
	}
)

func newContext() interface{} {
	h := new(HttpContext)
	h.proxyRequests = make([]http_service.IProxy, 0, 5)
	return h
}

package http_context

import (
	"sync"

	http_service "github.com/eolinker/eosc/eocontext/http-context"
)

var (
	pool = sync.Pool{
		New: newContext,
	}
	copyPool = sync.Pool{
		New: newCopyContext,
	}
)

func newContext() interface{} {
	h := new(HttpContext)
	h.proxyRequests = make([]http_service.IProxy, 0, 5)
	return h
}

func newCopyContext() interface{} {
	h := new(cloneContext)
	h.proxyRequests = make([]http_service.IProxy, 0, 5)
	return h
}

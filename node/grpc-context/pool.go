package grpc_context

import (
	"sync"
)

var (
	pool = sync.Pool{
		New: newContext,
	}
)

func newContext() interface{} {
	h := new(Context)
	return h
}

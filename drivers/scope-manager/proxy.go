package scope_manager

import (
	"sync/atomic"
)

var _ IProxy = (*Proxy)(nil)

type Proxy struct {
	pointer atomic.Pointer[[]interface{}]
}

func NewProxy() *Proxy {
	return &Proxy{pointer: atomic.Pointer[[]interface{}]{}}
}

func (p *Proxy) Set(values []interface{}) {
	p.pointer.Store(&values)
}

func (p *Proxy) List() []interface{} {
	t := p.pointer.Load()
	return *t
}

type IProxy interface {
	Set(values []interface{})
	IProxyOutput
}

type IProxyOutput interface {
	List() []interface{}
}

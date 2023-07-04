package scope_manager

import (
	"sync/atomic"
)

type _Proxy struct {
	pointer atomic.Pointer[[]interface{}]
}

func newProxy() *_Proxy {
	return &_Proxy{pointer: atomic.Pointer[[]interface{}]{}}
}

func (p *_Proxy) Set(values []interface{}) {
	p.pointer.Store(&values)
}

func (p *_Proxy) List() []interface{} {
	t := p.pointer.Load()
	if t == nil {
		return nil
	}
	return *t
}

type IProxy[T any] interface {
	Set(values ...T)
	IProxyOutput[T]
}

type IProxyOutput[T any] interface {
	List() []T
}

type Proxy[T any] struct {
	org *_Proxy
}
type StaticProxy[T any] struct {
	target []T
}

func (s *StaticProxy[T]) List() []T {
	return s.target
}

func NewProxy[T any](t ...T) IProxyOutput[T] {
	return &StaticProxy[T]{target: t}
}
func create[T any](proxy *_Proxy) IProxy[T] {
	return &Proxy[T]{
		org: proxy,
	}
}
func (p *Proxy[T]) Set(values ...T) {
	vs := make([]interface{}, 0, len(values))
	for _, v := range values {
		vs = append(vs, v)
	}
	p.org.Set(vs)
}

func (p *Proxy[T]) List() []T {
	values := p.org.List()
	vs := make([]T, 0, len(values))
	for _, v := range values {
		if vt, ok := v.(T); ok {
			vs = append(vs, vt)
		}
	}
	return vs
}

package http_entry

import (
	http_service "github.com/eolinker/eosc/eocontext/http-context"
)

type IProxyReader interface {
	ReadProxy(name string, proxy http_service.IProxy) (string, bool)
}

type ProxyReadFunc func(name string, proxy http_service.IProxy) (string, bool)

func (p ProxyReadFunc) ReadProxy(name string, proxy http_service.IProxy) (string, bool) {
	return p(name, proxy)
}

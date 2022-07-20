package http_entry

import (
	http_service "github.com/eolinker/eosc/context/http-context"
)

type IProxyReader interface {
	ReadProxy(name string, proxy http_service.IRequest) (string, bool)
}

type ProxyReadFunc func(name string, proxy http_service.IRequest) (string, bool)

func (p ProxyReadFunc) ReadProxy(name string, proxy http_service.IRequest) (string, bool) {
	return p(name, proxy)
}

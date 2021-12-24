package http_entry

import (
	http_service "github.com/eolinker/eosc/http-service"
)

type IReaderIndex interface {
	ReadByIndex(index int, name string, ctx http_service.IHttpContext) (string, bool)
}

type ProxyReaders map[string]IProxyReader

func (p ProxyReaders) ReadByIndex(index int, name string, ctx http_service.IHttpContext) (string, bool) {
	v, ok := p[name]
	if !ok {
		return "", false
	}
	proxies := ctx.Proxies()
	proxyLen := len(proxies)

	if proxyLen <= index {
		return "", false
	}
	if index == -1 {
		index = proxyLen - 1
	}

	return v.ReadProxy(name, proxies[index])

}

func (p ProxyReaders) Read(name string, ctx http_service.IHttpContext) (string, bool) {
	v, ok := p[name]
	if !ok {
		return "", false
	}
	proxies := ctx.Proxies()
	proxyLen := len(proxies)
	if proxyLen == 0 {
		return "", false
	}
	return v.ReadProxy(name, proxies[proxyLen-1])
}

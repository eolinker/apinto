package http_entry

import (
	"strings"

	http_service "github.com/eolinker/eosc/eocontext/http-context"
)

type IReaderIndex interface {
	ReadByIndex(index int, name string, ctx http_service.IHttpContext) (interface{}, bool)
}

type ProxyReaders map[string]IProxyReader

func (p ProxyReaders) ReadByIndex(index int, name string, ctx http_service.IHttpContext) (interface{}, bool) {
	proxies := ctx.Proxies()
	proxyLen := len(proxies)

	if proxyLen <= index {
		return "", false
	}
	if index == -1 {
		index = proxyLen - 1
	}
	v, ok := p[name]
	if !ok {
		fs := strings.SplitN(name, "_", 2)
		if len(fs) == 2 {
			v, ok = p[fs[0]]
			if ok {
				return v.ReadProxy(fs[1], proxies[index])
			}
		}
		return "", false
	}
	return v.ReadProxy("", proxies[index])
}

func (p ProxyReaders) Read(name string, ctx http_service.IHttpContext) (interface{}, bool) {
	ns := strings.SplitN(name, "_", 2)
	v, ok := p[ns[0]]
	if !ok {
		return "", false
	}
	proxies := ctx.Proxies()
	proxyLen := len(proxies)
	if proxyLen == 0 {
		return "", false
	}
	if len(ns) > 1 {
		return v.ReadProxy(ns[1], proxies[proxyLen-1])
	}
	return v.ReadProxy("", proxies[proxyLen-1])
}

package http_entry

import (
	"strings"

	http_service "github.com/eolinker/eosc/eocontext/http-context"
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

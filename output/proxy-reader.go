package output

import (
	"strings"

	http_service "github.com/eolinker/eosc/http-service"
)

type IProxyReader interface {
	Read(name string, proxy http_service.IRequest) (string, bool)
}

type ProxyReadFunc func(name string, proxy http_service.IRequest) (string, bool)

func (p ProxyReadFunc) Read(name string, proxy http_service.IRequest) (string, bool) {
	return p(name, proxy)
}

type Proxies map[string]IProxyReader

func (p Proxies) Read(name string, proxy http_service.IRequest) (string, bool) {
	r, has := p[name]
	if has {
		return r.Read("", proxy)
	}
	fs := strings.SplitN(name, "_", 2)
	if len(fs) != 2 {
		return r.Read("", proxy)
	}
	r, has = p[fs[0]]
	if has {
		return r.Read(fs[1], proxy)
	}
	return "", false
}

var (
	proxyFields Proxies = map[string]IProxyReader{
		"header": ProxyReadFunc(func(name string, proxy http_service.IRequest) (string, bool) {
			if name == "" {
				return proxy.Header().RawHeader(), true
			}
			return proxy.Header().GetHeader(name), true
		}),
		"uri": ProxyReadFunc(func(name string, proxy http_service.IRequest) (string, bool) {
			return proxy.URI().RawURL(), true
		}),
		"body": ProxyReadFunc(func(name string, proxy http_service.IRequest) (string, bool) {
			body, err := proxy.Body().RawBody()
			if err != nil {
				return "", false
			}
			return string(body), true
		}),
		"addr": ProxyReadFunc(func(name string, proxy http_service.IRequest) (string, bool) {
			return proxy.URI().Host(), true
		}),
		"scheme": ProxyReadFunc(func(name string, proxy http_service.IRequest) (string, bool) {
			return proxy.URI().Scheme(), true
		}),
		"method": ProxyReadFunc(func(name string, proxy http_service.IRequest) (string, bool) {
			return proxy.Method(), true
		}),
	}
)

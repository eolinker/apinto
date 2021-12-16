package http_entry

import (
	"strings"

	http_service "github.com/eolinker/eosc/http-service"
)

type IReader interface {
	Read(name string, ctx http_service.IHttpContext) (string, bool)
}

type ReadFunc func(name string, ctx http_service.IHttpContext) (string, bool)

func (f ReadFunc) Read(name string, ctx http_service.IHttpContext) (string, bool) {
	return f(name, ctx)
}

type Fields map[string]IReader

func (f Fields) Read(name string, ctx http_service.IHttpContext) (string, bool) {
	r, has := f[name]
	if has {
		return r.Read("", ctx)
	}
	fs := strings.SplitN(name, "_", 2)
	if len(fs) != 2 {
		return r.Read("", ctx)
	}
	r, has = f[fs[0]]
	if has {
		return r.Read(fs[1], ctx)
	}
	return "", false
}

var (
	rule Fields = map[string]IReader{
		"request_id": ReadFunc(func(name string, ctx http_service.IHttpContext) (string, bool) {
			return ctx.RequestId(), true
		}),
		"content_length": ReadFunc(func(name string, ctx http_service.IHttpContext) (string, bool) {
			return ctx.Request().Header().GetHeader("content-length"), true
		}),
		"content_type": ReadFunc(func(name string, ctx http_service.IHttpContext) (string, bool) {
			return ctx.Request().Header().GetHeader("content-type"), true
		}),
		"http": ReadFunc(func(name string, ctx http_service.IHttpContext) (string, bool) {
			return ctx.Request().Header().GetHeader(strings.Replace(name, "_", "-", -1)), true
		}),
		"proxy": proxyFields,
	}

	proxyFields = ProxyReaders{
		"header": ProxyReadFunc(func(name string, proxy http_service.IRequest) (string, bool) {
			if name == "" {
				return proxy.Header().RawHeader(), true
			}
			return proxy.Header().GetHeader(strings.Replace(name, "_", "-", -1)), true
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

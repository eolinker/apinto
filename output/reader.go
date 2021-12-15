package output

import (
	"strings"

	http_service "github.com/eolinker/eosc/http-service"
)

type IReader interface {
	Read(name string, index int, ctx http_service.IHttpContext) (string, bool)
}

type ReadFunc func(name string, index int, ctx http_service.IHttpContext) (string, bool)

func (f ReadFunc) Read(name string, index int, ctx http_service.IHttpContext) (string, bool) {
	return f(name, index, ctx)
}

type Fields map[string]IReader

func (f Fields) Read(name string, index int, ctx http_service.IHttpContext) (string, bool) {
	r, has := f[name]
	if has {
		return r.Read("", index, ctx)
	}
	fs := strings.SplitN(name, "_", 2)
	if len(fs) != 2 {
		return r.Read("", index, ctx)
	}
	r, has = f[fs[0]]
	if has {
		return r.Read(fs[1], index, ctx)
	}
	return "", false
}

var (
	rule Fields = map[string]IReader{
		"request_id": ReadFunc(func(name string, index int, ctx http_service.IHttpContext) (string, bool) {
			return ctx.RequestId(), true
		}),
		"content_length": ReadFunc(func(name string, index int, ctx http_service.IHttpContext) (string, bool) {
			return ctx.Request().Header().GetHeader("content-length"), true
		}),
		"content_type": ReadFunc(func(name string, index int, ctx http_service.IHttpContext) (string, bool) {
			return ctx.Request().Header().GetHeader("content-type"), true
		}),
		"http": ReadFunc(func(name string, index int, ctx http_service.IHttpContext) (string, bool) {
			return ctx.Request().Header().GetHeader(name), true
		}),
		"proxy": ReadFunc(func(name string, index int, ctx http_service.IHttpContext) (string, bool) {
			proxies := ctx.Proxies()
			proxyLen := len(proxies)

			if proxyLen <= index {
				return "", false
			}
			if index == -1 {
				index = proxyLen - 1
			}
			v, ok := proxyFields[name]
			if ok {
				return v.Read(name, proxies[index])
			}
			return "", false
		}),
	}
)

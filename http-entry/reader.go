package http_entry

import (
	"fmt"
	"strings"
	"time"

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
		"query": ReadFunc(func(name string, ctx http_service.IHttpContext) (string, bool) {
			//TODO query需要返回完整的请求参数吗
			v := ctx.Request().URI().GetQuery(name)
			if v == "" {
				return "", false
			}
			return v, true
		}),
		"uri": ReadFunc(func(name string, ctx http_service.IHttpContext) (string, bool) {
			//不带请求参数的uri
			return ctx.Request().URI().Path(), true
		}),
		"content_length": ReadFunc(func(name string, ctx http_service.IHttpContext) (string, bool) {
			return ctx.Request().Header().GetHeader("content-length"), true
		}),
		"content_type": ReadFunc(func(name string, ctx http_service.IHttpContext) (string, bool) {
			return ctx.Request().Header().GetHeader("content-type"), true
		}),
		"cookie": ReadFunc(func(name string, ctx http_service.IHttpContext) (string, bool) {
			//TODO
			if name == "" {
				return "完整cookie", true
			}
			name = strings.Replace(name, "_", "-", -1)
			return "", true
		}),
		"msec": ReadFunc(func(name string, ctx http_service.IHttpContext) (string, bool) {
			return fmt.Sprintf("%d", time.Now().Unix()), true
		}),
		"apinto_version": ReadFunc(func(name string, ctx http_service.IHttpContext) (string, bool) {
			//TODO
			return "", true
		}),
		"remote_addr": ReadFunc(func(name string, ctx http_service.IHttpContext) (string, bool) {
			//是ip地址还是整个地址
			return ctx.Request().RemoteAddr(), true
		}),
		"remote_port": ReadFunc(func(name string, ctx http_service.IHttpContext) (string, bool) {
			//TODO 需要从fasthttpContext 里面获取RemoteAddr
			return "", true
		}),
		"request": ReadFunc(func(name string, ctx http_service.IHttpContext) (string, bool) {

			method := ctx.Request().Method()
			uri := ctx.Request().URI().RequestURI()
			//TODO 获取的不包含/1.1, 怎么处理？
			proto := ctx.Request().URI().Scheme()
			return fmt.Sprintf("%s %s %s", method, uri, proto), true
		}),
		"request_body": ReadFunc(func(name string, ctx http_service.IHttpContext) (string, bool) {
			body, err := ctx.Request().Body().RawBody()
			if err != nil {
				return "", false
			}
			return string(body), true
		}),
		"request_length": ReadFunc(func(name string, ctx http_service.IHttpContext) (string, bool) {
			//包括请求的地址，http请求头和请求主体
			uriLen := len(ctx.Request().URI().RequestURI())
			headerLen := len(ctx.Request().Header().RawHeader())
			body, err := ctx.Request().Body().RawBody()
			//TODO 返回false还是 bodyLen为0
			if err != nil {
				return "", false
			}
			bodyLen := len(body)
			return fmt.Sprint(uriLen + headerLen + bodyLen), true
		}),
		"request_method": ReadFunc(func(name string, ctx http_service.IHttpContext) (string, bool) {
			return ctx.Request().Method(), true
		}),
		"request_time": ReadFunc(func(name string, ctx http_service.IHttpContext) (string, bool) {
			//TODO

			return "", true
		}),
		"request_uri": ReadFunc(func(name string, ctx http_service.IHttpContext) (string, bool) {
			return ctx.Request().URI().RequestURI(), true
		}),
		"scheme": ReadFunc(func(name string, ctx http_service.IHttpContext) (string, bool) {
			return ctx.Request().URI().Scheme(), true
		}),
		"status": ReadFunc(func(name string, ctx http_service.IHttpContext) (string, bool) {
			return ctx.Response().Status(), true
		}),
		"time_iso8601": ReadFunc(func(name string, ctx http_service.IHttpContext) (string, bool) {
			//带毫秒的ISO-8601时间格式
			return time.Now().Format("2006-01-02T15:04:05.000Z07:00"), true
		}),
		"time_local": ReadFunc(func(name string, ctx http_service.IHttpContext) (string, bool) {
			return time.Now().Format("2006-01-02 15:04:05"), true
		}),
		"header": ReadFunc(func(name string, ctx http_service.IHttpContext) (string, bool) {
			return ctx.Request().Header().RawHeader(), true
		}),
		"http": ReadFunc(func(name string, ctx http_service.IHttpContext) (string, bool) {
			return ctx.Request().Header().GetHeader(strings.Replace(name, "_", "-", -1)), true
		}),
		"host": ReadFunc(func(name string, ctx http_service.IHttpContext) (string, bool) {
			return fmt.Sprintf("%d", time.Now().Unix()), true
		}),
		"error": ReadFunc(func(name string, ctx http_service.IHttpContext) (string, bool) {
			//TODO 暂时忽略
			return "", true
		}),
		"response": ReadFunc(func(name string, ctx http_service.IHttpContext) (string, bool) {
			//TODO
			return fmt.Sprintf("%d", time.Now().Unix()), true
		}),
		"response_body": ReadFunc(func(name string, ctx http_service.IHttpContext) (string, bool) {
			return string(ctx.Response().GetBody()), true
		}),
		"response_header": ReadFunc(func(name string, ctx http_service.IHttpContext) (string, bool) {
			//TODO
			return "", true
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

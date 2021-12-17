package http_entry

import (
	"fmt"
	"github.com/eolinker/goku/utils"
	"strconv"
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
			if name == "" {
				return ctx.Request().URI().RawQuery(), true
			}
			//TODO 返回的布尔值问题
			value := ctx.Request().URI().GetQuery(name)
			if value == "" {
				return "", false
			}
			return value, true
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
			if name == "" {
				return ctx.Request().Header().GetHeader("cookie"), true
			}
			//TODO
			name = strings.Replace(name, "_", "-", -1)
			return "", true
		}),
		"msec": ReadFunc(func(name string, ctx http_service.IHttpContext) (string, bool) {
			return strconv.FormatInt(time.Now().Unix(), 10), true
		}),
		"apinto_version": ReadFunc(func(name string, ctx http_service.IHttpContext) (string, bool) {
			return utils.Version, true
		}),
		"remote_addr": ReadFunc(func(name string, ctx http_service.IHttpContext) (string, bool) {
			return ctx.Request().RemoteAddr(), true
		}),
		"remote_port": ReadFunc(func(name string, ctx http_service.IHttpContext) (string, bool) {
			return ctx.Request().RemotePort(), false
		}),
		"request": Fields{
			"": ReadFunc(func(name string, ctx http_service.IHttpContext) (string, bool) {
				//TODO
				ctx.Request().String()
				return " ", true
			}),
			"body": ReadFunc(func(name string, ctx http_service.IHttpContext) (string, bool) {
				body, err := ctx.Request().Body().RawBody()
				if err != nil {
					return "", false
				}
				return string(body), true
			}),
			"length": ReadFunc(func(name string, ctx http_service.IHttpContext) (string, bool) {

				return strconv.Itoa(len(ctx.Request().String())), true
			}),
			"method": ReadFunc(func(name string, ctx http_service.IHttpContext) (string, bool) {
				return ctx.Request().Method(), true
			}),
			"time": ReadFunc(func(name string, ctx http_service.IHttpContext) (string, bool) {
				//TODO
				requestTime := ctx.Value("request_time")
				start, ok := requestTime.(time.Time)
				if !ok {
					return "", false
				}
				_ = time.Now().Sub(start).Seconds()
				return "", true
			}),
			"uri": ReadFunc(func(name string, ctx http_service.IHttpContext) (string, bool) {
				return ctx.Request().URI().RequestURI(), true
			}),
		},
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
		"response": Fields{
			"": ReadFunc(func(name string, ctx http_service.IHttpContext) (string, bool) {
				return ctx.Response().String(), true
			}),
			"body": ReadFunc(func(name string, ctx http_service.IHttpContext) (string, bool) {
				ctx.Response().GetBody()
				return fmt.Sprintf("%d", time.Now().Unix()), true
			}),
			"header": ReadFunc(func(name string, ctx http_service.IHttpContext) (string, bool) {
				if name == "" {
					return ctx.Response().HeadersString(), true
				}
				return ctx.Response().GetHeader(strings.Replace(name, "_", "-", -1)), true
			}),
		},
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

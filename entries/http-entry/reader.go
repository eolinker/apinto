package http_entry

import (
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/eolinker/apinto/utils/version"

	"github.com/eolinker/apinto/utils"
	http_service "github.com/eolinker/eosc/eocontext/http-context"
)

type IReader interface {
	Read(name string, ctx http_service.IHttpContext) (interface{}, bool)
}

type ReadFunc func(name string, ctx http_service.IHttpContext) (interface{}, bool)

func (f ReadFunc) Read(name string, ctx http_service.IHttpContext) (interface{}, bool) {
	return f(name, ctx)
}

type Fields map[string]IReader

func (f Fields) Read(name string, ctx http_service.IHttpContext) (interface{}, bool) {
	r, has := f[name]
	if has {
		return r.Read("", ctx)
	}
	fs := strings.SplitN(name, "_", 2)
	if len(fs) == 2 {
		r, has = f[fs[0]]
		if has {
			return r.Read(fs[1], ctx)
		}
	}

	label := ctx.GetLabel(name)
	if label != "" {
		return label, true
	}
	label = os.Getenv(name)

	return label, label != ""
}

var (
	rule Fields = map[string]IReader{
		"request_id": ReadFunc(func(name string, ctx http_service.IHttpContext) (interface{}, bool) {
			return ctx.RequestId(), true
		}),
		"node": ReadFunc(func(name string, ctx http_service.IHttpContext) (interface{}, bool) {
			return os.Getenv("node_id"), true
		}),
		"cluster": ReadFunc(func(name string, ctx http_service.IHttpContext) (interface{}, bool) {
			return os.Getenv("cluster_id"), true
		}),
		"api_id": ReadFunc(func(name string, ctx http_service.IHttpContext) (interface{}, bool) {
			return ctx.GetLabel("api_id"), true
		}),
		"query": ReadFunc(func(name string, ctx http_service.IHttpContext) (interface{}, bool) {
			if name == "" {
				return utils.QueryUrlEncode(ctx.Request().URI().RawQuery()), true
			}
			return url.QueryEscape(ctx.Request().URI().GetQuery(name)), true
		}),
		"uri": ReadFunc(func(name string, ctx http_service.IHttpContext) (interface{}, bool) {
			//不带请求参数的uri
			return ctx.Request().URI().Path(), true
		}),
		"content_length": ReadFunc(func(name string, ctx http_service.IHttpContext) (interface{}, bool) {
			return ctx.Request().Header().GetHeader("content-length"), true
		}),
		"content_type": ReadFunc(func(name string, ctx http_service.IHttpContext) (interface{}, bool) {
			return ctx.Request().Header().GetHeader("content-type"), true
		}),
		"cookie": ReadFunc(func(name string, ctx http_service.IHttpContext) (interface{}, bool) {
			if name == "" {
				return ctx.Request().Header().GetHeader("cookie"), true
			}
			return ctx.Request().Header().GetCookie(name), false
		}),
		"msec": ReadFunc(func(name string, ctx http_service.IHttpContext) (interface{}, bool) {
			return ctx.AcceptTime().UnixMilli(), true
		}),
		"apinto_version": ReadFunc(func(name string, ctx http_service.IHttpContext) (interface{}, bool) {
			return version.Version, true
		}),
		"remote_addr": ReadFunc(func(name string, ctx http_service.IHttpContext) (interface{}, bool) {
			return ctx.Request().RemoteAddr(), true
		}),
		"remote_port": ReadFunc(func(name string, ctx http_service.IHttpContext) (interface{}, bool) {
			return ctx.Request().RemotePort(), true
		}),

		"request": Fields{
			"body": ReadFunc(func(name string, ctx http_service.IHttpContext) (interface{}, bool) {
				body, err := ctx.Request().Body().RawBody()
				if err != nil {
					return "", false
				}
				return string(body), true
			}),
			"length": ReadFunc(func(name string, ctx http_service.IHttpContext) (interface{}, bool) {

				return ctx.Request().ContentLength(), true
			}),
			"method": ReadFunc(func(name string, ctx http_service.IHttpContext) (interface{}, bool) {
				return ctx.Request().Method(), true
			}),
			"time": ReadFunc(func(name string, ctx http_service.IHttpContext) (interface{}, bool) {
				requestTime := ctx.Value("request_time")
				start, ok := requestTime.(time.Time)
				if !ok {
					return "", false
				}
				return time.Now().Sub(start).Milliseconds(), true
			}),
			"uri": ReadFunc(func(name string, ctx http_service.IHttpContext) (interface{}, bool) {
				return ctx.Request().URI().RequestURI(), true
			}),
		},

		"scheme": ReadFunc(func(name string, ctx http_service.IHttpContext) (interface{}, bool) {
			return ctx.Request().URI().Scheme(), true
		}),
		"status": ReadFunc(func(name string, ctx http_service.IHttpContext) (interface{}, bool) {
			return ctx.Response().StatusCode(), true
		}),
		"time_iso8601": ReadFunc(func(name string, ctx http_service.IHttpContext) (interface{}, bool) {
			//带毫秒的ISO-8601时间格式
			//return time.Now().Format("2006-01-02T15:04:05.000Z07:00"), true
			return ctx.AcceptTime().Format("2006-01-02T15:04:05.000Z07:00"), true
		}),
		"time_local": ReadFunc(func(name string, ctx http_service.IHttpContext) (interface{}, bool) {
			//return time.Now().Format("2006-01-02 15:04:05"), true
			return ctx.AcceptTime().Format("2006-01-02 15:04:05"), true
		}),
		"header": ReadFunc(func(name string, ctx http_service.IHttpContext) (interface{}, bool) {
			if name == "" {
				return url.Values(ctx.Request().Header().Headers()).Encode(), true
			}
			return ctx.Request().Header().GetHeader(strings.Replace(name, "_", "-", -1)), true
		}),
		"http": ReadFunc(func(name string, ctx http_service.IHttpContext) (interface{}, bool) {
			return ctx.Request().Header().GetHeader(strings.Replace(name, "_", "-", -1)), true
		}),
		"host": ReadFunc(func(name string, ctx http_service.IHttpContext) (interface{}, bool) {
			return ctx.Request().URI().Host(), true
		}),
		"error": ReadFunc(func(name string, ctx http_service.IHttpContext) (interface{}, bool) {
			//TODO 暂时忽略
			return "", true
		}),

		"response": Fields{
			"": ReadFunc(func(name string, ctx http_service.IHttpContext) (interface{}, bool) {
				return ctx.Response().String(), true
			}),
			"body": ReadFunc(func(name string, ctx http_service.IHttpContext) (interface{}, bool) {
				return string(ctx.Response().GetBody()), true
			}),
			"header": ReadFunc(func(name string, ctx http_service.IHttpContext) (interface{}, bool) {
				if name == "" {
					return url.Values(ctx.Response().Headers()).Encode(), true
				}
				return ctx.Response().GetHeader(strings.Replace(name, "_", "-", -1)), true
			}),
			"status": ReadFunc(func(name string, ctx http_service.IHttpContext) (interface{}, bool) {
				return ctx.Response().ProxyStatus(), true
			}),
			"time": ReadFunc(func(name string, ctx http_service.IHttpContext) (interface{}, bool) {
				return ctx.Response().ResponseTime(), true
			}),
			"length": ReadFunc(func(name string, ctx http_service.IHttpContext) (interface{}, bool) {
				return strconv.Itoa(ctx.Response().ContentLength()), true
			}),
		},
		"proxy": proxyFields,
	}

	proxyFields = ProxyReaders{
		"header": ProxyReadFunc(func(name string, proxy http_service.IProxy) (interface{}, bool) {
			if name == "" {
				return url.Values(proxy.Header().Headers()).Encode(), true
			}

			return proxy.Header().GetHeader(strings.Replace(name, "_", "-", -1)), true
		}),
		"uri": ProxyReadFunc(func(name string, proxy http_service.IProxy) (interface{}, bool) {
			return proxy.URI().RequestURI(), true
		}),
		"query": ProxyReadFunc(func(name string, proxy http_service.IProxy) (interface{}, bool) {
			if name == "" {
				return utils.QueryUrlEncode(proxy.URI().RawQuery()), true
			}
			return url.QueryEscape(proxy.URI().GetQuery(name)), true
		}),
		"body": ProxyReadFunc(func(name string, proxy http_service.IProxy) (interface{}, bool) {
			body, err := proxy.Body().RawBody()
			if err != nil {
				return "", false
			}
			return string(body), true
		}),
		"addr": ProxyReadFunc(func(name string, proxy http_service.IProxy) (interface{}, bool) {
			return proxy.URI().Host(), true
		}),
		"scheme": ProxyReadFunc(func(name string, proxy http_service.IProxy) (interface{}, bool) {
			return proxy.URI().Scheme(), true
		}),
		"method": ProxyReadFunc(func(name string, proxy http_service.IProxy) (interface{}, bool) {
			return proxy.Method(), true
		}),
		"status": ProxyReadFunc(func(name string, proxy http_service.IProxy) (interface{}, bool) {
			return proxy.StatusCode(), true
		}),
		"path": ProxyReadFunc(func(name string, proxy http_service.IProxy) (interface{}, bool) {
			return proxy.URI().Path(), true
		}),
		"host": ProxyReadFunc(func(name string, proxy http_service.IProxy) (interface{}, bool) {
			return proxy.Header().Host(), true
		}),
		"request_length": ProxyReadFunc(func(name string, proxy http_service.IProxy) (interface{}, bool) {
			return proxy.ContentLength(), true
		}),
		"response_length": ProxyReadFunc(func(name string, proxy http_service.IProxy) (interface{}, bool) {
			return proxy.ResponseLength(), true
		}),
		"response_body": ProxyReadFunc(func(name string, proxy http_service.IProxy) (interface{}, bool) {
			return proxy.ResponseBody(), true
		}),
		"time": ProxyReadFunc(func(name string, proxy http_service.IProxy) (interface{}, bool) {
			return proxy.ResponseTime(), true
		}),
		"msec": ProxyReadFunc(func(name string, proxy http_service.IProxy) (interface{}, bool) {
			return proxy.ProxyTime().UnixMilli(), true
		}),
	}
)

func GetProxyReaders() ProxyReaders {
	return proxyFields
}

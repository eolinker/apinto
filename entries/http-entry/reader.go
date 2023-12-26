package http_entry

import (
	"fmt"
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

type CtxRule struct {
	fields Fields
}

func (l *CtxRule) Read(name string, ctx http_service.IHttpContext) (interface{}, bool) {
	value := ctx.Value(name)
	if value != nil {
		return value, true
	}
	// 先从Label中获取值
	value = ctx.GetLabel(name)
	if value != "" {
		return value, true
	}

	return l.fields.Read(name, ctx)
}
func init() {
	ctxRule = &CtxRule{
		fields: rule,
	}
}

var (
	ctxRule *CtxRule
	rule    Fields = map[string]IReader{
		"request_id": ReadFunc(func(name string, ctx http_service.IHttpContext) (interface{}, bool) {
			return ctx.RequestId(), true
		}),
		"node": ReadFunc(func(name string, ctx http_service.IHttpContext) (interface{}, bool) {
			return os.Getenv("node_id"), true
		}),
		"cluster": ReadFunc(func(name string, ctx http_service.IHttpContext) (interface{}, bool) {
			return os.Getenv("cluster_id"), true
		}),
		"query": ReadFunc(func(name string, ctx http_service.IHttpContext) (interface{}, bool) {
			if name == "" {
				return utils.QueryUrlEncode(ctx.Request().URI().RawQuery()), true
			}
			return url.QueryEscape(ctx.Request().URI().GetQuery(name)), true
		}),
		"src_ip": ReadFunc(func(name string, ctx http_service.IHttpContext) (interface{}, bool) {
			return ctx.Request().RealIp(), true
		}),
		"src_port": ReadFunc(func(name string, ctx http_service.IHttpContext) (interface{}, bool) {
			port, err := strconv.Atoi(ctx.Request().RemotePort())
			if err != nil {
				return nil, false
			}
			return port, true
		}),
		"uri": ReadFunc(func(name string, ctx http_service.IHttpContext) (interface{}, bool) {
			//不带请求参数的uri
			return ctx.Request().URI().Path(), true
		}),
		"url": ReadFunc(func(name string, ctx http_service.IHttpContext) (interface{}, bool) {
			return ctx.Request().URI().RawURL(), true
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
		"ctx": ReadFunc(func(name string, ctx http_service.IHttpContext) (interface{}, bool) {
			return ctxRule.Read(name, ctx)

		}),

		"request": Fields{
			"body": ReadFunc(func(name string, ctx http_service.IHttpContext) (interface{}, bool) {
				body, err := ctx.Request().Body().RawBody()
				if err != nil {
					return "", false
				}
				return string(body), true
			}),
			"body_filter": ReadFunc(func(name string, ctx http_service.IHttpContext) (interface{}, bool) {
				value := ctx.GetLabel("xxx")
				if value == "" {
					body, err := ctx.Request().Body().RawBody()
					if err != nil {
						return "", false
					}
					return string(body), true
				}
				return ctx.GetLabel("xxx"), true
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
		"timestamp": ReadFunc(func(name string, ctx http_service.IHttpContext) (interface{}, bool) {
			return ctx.AcceptTime().Unix(), true
		}),
		"header": ReadFunc(func(name string, ctx http_service.IHttpContext) (interface{}, bool) {
			if name == "" {
				return url.Values(ctx.Request().Header().Headers()).Encode(), true
			}
			return ctx.Request().Header().GetHeader(strings.Replace(name, "_", "-", -1)), true
		}),
		"headers": ReadFunc(func(name string, ctx http_service.IHttpContext) (interface{}, bool) {
			result := make(map[string]string)
			for key, value := range ctx.Request().Header().Headers() {
				result[strings.ToLower(key)] = strings.Join(value, ";")
			}
			return result, true
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
			"body": Fields{
				"": ReadFunc(func(name string, ctx http_service.IHttpContext) (interface{}, bool) {
					return string(ctx.Response().GetBody()), true
				}),
			},
			"header": ReadFunc(func(name string, ctx http_service.IHttpContext) (interface{}, bool) {
				if name == "" {
					return url.Values(ctx.Response().Headers()).Encode(), true
				}
				return ctx.Response().GetHeader(strings.Replace(name, "_", "-", -1)), true
			}),
			"headers": ReadFunc(func(name string, ctx http_service.IHttpContext) (interface{}, bool) {
				result := make(map[string]string)
				for key, value := range ctx.Response().Headers() {
					result[strings.ToLower(key)] = strings.Join(value, ";")
				}
				return result, true
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
		"set_cookies": ReadFunc(func(name string, ctx http_service.IHttpContext) (interface{}, bool) {
			cookies := ctx.Response().GetHeader("Set-Cookie")
			if strings.TrimSpace(cookies) == "" {
				return nil, true
			}
			return strings.Split(ctx.Response().GetHeader("Set-Cookie"), "; "), true
		}),
		"dst_ip": ReadFunc(func(name string, ctx http_service.IHttpContext) (interface{}, bool) {
			return ctx.Response().RemoteIP(), true
		}),
		"dst_port": ReadFunc(func(name string, ctx http_service.IHttpContext) (interface{}, bool) {
			return ctx.Response().RemotePort(), true
		}),
		"proxy": proxyFields,
	}

	proxyFields = ProxyReaders{
		"header": ProxyReadFunc(func(name string, proxy http_service.IProxy) (interface{}, bool) {
			if name == "" {
				return url.Values(proxy.Header().Headers()).Encode(), true
			}

			return proxy.Header().GetHeader(strings.Replace(name, "_", "-", -1)), true
		}),
		"headers": ProxyReadFunc(func(name string, proxy http_service.IProxy) (interface{}, bool) {
			result := make(map[string]string)
			for key, value := range proxy.Header().Headers() {
				result[strings.ToLower(key)] = strings.Join(value, ";")
			}
			return result, true
		}),
		"uri": ProxyReadFunc(func(name string, proxy http_service.IProxy) (interface{}, bool) {
			return proxy.URI().RequestURI(), true
		}),
		"url": ProxyReadFunc(func(name string, proxy http_service.IProxy) (interface{}, bool) {
			return fmt.Sprintf("%s://%s:%d%s", proxy.URI().Scheme(), proxy.RemoteIP(), proxy.RemotePort(), proxy.URI().RequestURI()), true
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
		"dst_ip": ProxyReadFunc(func(name string, proxy http_service.IProxy) (interface{}, bool) {
			return proxy.RemoteIP(), true
		}),
		"dst_port": ProxyReadFunc(func(name string, proxy http_service.IProxy) (interface{}, bool) {
			return proxy.RemotePort(), true
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
		"response_headers": ProxyReadFunc(func(name string, proxy http_service.IProxy) (interface{}, bool) {
			result := make(map[string]string)
			for key, value := range proxy.ResponseHeaders() {
				result[strings.ToLower(key)] = strings.Join(value, ";")
			}
			return result, true
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

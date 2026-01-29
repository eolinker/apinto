package http_entry

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"github.com/eolinker/eosc/log"
	"io"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/eolinker/eosc/env"

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
	value := ctx.Value(name)
	if value != nil {
		return value, true
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
		"gateway_host": ReadFunc(func(name string, ctx http_service.IHttpContext) (interface{}, bool) {
			hosts, has := env.GetEnv("GATEWAY_ADVERTISE_HOSTS")
			if !has {
				return "", false
			}
			return strings.Split(hosts, ",")[0], true
		}),
		"gateway_hosts": ReadFunc(func(name string, ctx http_service.IHttpContext) (interface{}, bool) {
			hosts, has := env.GetEnv("GATEWAY_ADVERTISE_HOSTS")
			if !has {
				return "", false
			}
			return strings.Split(hosts, ","), true
		}),
		"peer_host": ReadFunc(func(name string, ctx http_service.IHttpContext) (interface{}, bool) {
			hosts, has := env.GetEnv("PEER_ADVERTISE_HOSTS")
			if !has {
				return "", false
			}
			return strings.Split(hosts, ",")[0], true
		}),
		"peer_hosts": ReadFunc(func(name string, ctx http_service.IHttpContext) (interface{}, bool) {
			hosts, has := env.GetEnv("PEER_ADVERTISE_HOSTS")
			if !has {
				return "", false
			}
			return strings.Split(hosts, ","), true
		}),
		"client_host": ReadFunc(func(name string, ctx http_service.IHttpContext) (interface{}, bool) {
			hosts, has := env.GetEnv("CLIENT_ADVERTISE_HOSTS")
			if !has {
				return "", false
			}
			return strings.Split(hosts, ",")[0], true
		}),
		"client_hosts": ReadFunc(func(name string, ctx http_service.IHttpContext) (interface{}, bool) {
			hosts, has := env.GetEnv("CLIENT_ADVERTISE_HOSTS")
			if !has {
				return "", false
			}
			return strings.Split(hosts, ","), true
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
				builder := &strings.Builder{}
				builder.WriteString("HTTP/1.1 ")
				builder.WriteString(strconv.Itoa(ctx.Response().StatusCode()))
				builder.WriteString(" ")
				builder.WriteString(ctx.Response().Status())
				builder.WriteString("\r\n")
				err := ctx.Response().Headers().Write(builder)
				if err != nil {
					log.Errorf("write response headers error: %v", err)
					return "", false
				}

				body := ctx.Response().GetBody()
				encoding := ctx.Response().Headers().Get("Content-Encoding")
				if encoding == "gzip" {
					reader, err := gzip.NewReader(bytes.NewReader(body))
					if err != nil {
						return "", false
					}
					defer reader.Close()
					data, err := io.ReadAll(reader)
					if err != nil {
						return "", false
					}
					body = data
				}
				_, err = builder.Write(body)
				if err != nil {
					log.Errorf("write response body error: %v", err)
					return "", false
				}
				return builder.String(), true
			}),
			"body": Fields{
				"": ReadFunc(func(name string, ctx http_service.IHttpContext) (interface{}, bool) {
					body := string(ctx.Response().GetBody())
					encoding := ctx.Response().Headers().Get("Content-Encoding")
					if encoding == "gzip" {
						reader, err := gzip.NewReader(bytes.NewReader([]byte(body)))
						if err != nil {
							return "", false
						}
						defer reader.Close()
						data, err := io.ReadAll(reader)
						if err != nil {
							return "", false
						}
						return string(data), true
					}
					return body, true
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
			cookies := ctx.Response().GetHeader("SetProvider-Cookie")
			if strings.TrimSpace(cookies) == "" {
				return nil, true
			}
			return strings.Split(ctx.Response().GetHeader("SetProvider-Cookie"), "; "), true
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
		"header": &proxyReader{
			ProxyReadFunc: ProxyReadFunc(func(name string, proxy http_service.IProxy) (interface{}, bool) {
				if name == "" {
					return url.Values(proxy.Header().Headers()).Encode(), true
				}

				return proxy.Header().GetHeader(strings.Replace(name, "_", "-", -1)), true
			}),
			ProxyReadRequestFunc: ProxyReadRequestFunc(func(name string, proxy http_service.IRequest) (interface{}, bool) {
				if name == "" {
					return url.Values(proxy.Header().Headers()).Encode(), true
				}
				return proxy.Header().GetHeader(strings.Replace(name, "_", "-", -1)), true
			}),
		},
		"headers": &proxyReader{
			ProxyReadFunc: ProxyReadFunc(func(name string, proxy http_service.IProxy) (interface{}, bool) {
				result := make(map[string]string)
				for key, value := range proxy.Header().Headers() {
					result[strings.ToLower(key)] = strings.Join(value, ";")
				}
				return result, true
			}),
			ProxyReadRequestFunc: ProxyReadRequestFunc(func(name string, proxy http_service.IRequest) (interface{}, bool) {
				result := make(map[string]string)
				for key, value := range proxy.Header().Headers() {
					result[strings.ToLower(key)] = strings.Join(value, ";")
				}
				return result, true
			}),
		},
		"uri": &proxyReader{
			ProxyReadFunc: ProxyReadFunc(func(name string, proxy http_service.IProxy) (interface{}, bool) {
				return proxy.URI().RequestURI(), true
			}),
			ProxyReadRequestFunc: ProxyReadRequestFunc(func(name string, proxy http_service.IRequest) (interface{}, bool) {
				return proxy.URI().RequestURI(), true
			}),
		},
		"url": &proxyReader{
			ProxyReadFunc: ProxyReadFunc(func(name string, proxy http_service.IProxy) (interface{}, bool) {
				return fmt.Sprintf("%s://%s:%d%s", proxy.URI().Scheme(), proxy.RemoteIP(), proxy.RemotePort(), proxy.URI().RequestURI()), true
			}),
			ProxyReadRequestFunc: ProxyReadRequestFunc(func(name string, proxy http_service.IRequest) (interface{}, bool) {
				return fmt.Sprintf("%s://%s:%d%s", proxy.URI().Scheme(), "unknown", 0, proxy.URI().RequestURI()), true
			}),
		},
		"query": &proxyReader{
			ProxyReadFunc: ProxyReadFunc(func(name string, proxy http_service.IProxy) (interface{}, bool) {
				if name == "" {
					return utils.QueryUrlEncode(proxy.URI().RawQuery()), true
				}
				return url.QueryEscape(proxy.URI().GetQuery(name)), true
			}),
			ProxyReadRequestFunc: ProxyReadRequestFunc(func(name string, proxy http_service.IRequest) (interface{}, bool) {
				if name == "" {
					return utils.QueryUrlEncode(proxy.URI().RawQuery()), true
				}
				return url.QueryEscape(proxy.URI().GetQuery(name)), true
			}),
		},
		"body": &proxyReader{
			ProxyReadFunc: ProxyReadFunc(func(name string, proxy http_service.IProxy) (interface{}, bool) {
				body, err := proxy.Body().RawBody()
				if err != nil {
					return "", false
				}
				return string(body), true
			}),
			ProxyReadRequestFunc: ProxyReadRequestFunc(func(name string, proxy http_service.IRequest) (interface{}, bool) {
				body, err := proxy.Body().RawBody()
				if err != nil {
					return "", false
				}
				return string(body), true
			}),
		},
		"addr": &proxyReader{
			ProxyReadFunc: ProxyReadFunc(func(name string, proxy http_service.IProxy) (interface{}, bool) {
				return proxy.URI().Host(), true
			}),
			ProxyReadRequestFunc: ProxyReadRequestFunc(func(name string, proxy http_service.IRequest) (interface{}, bool) {
				return proxy.URI().Host(), true
			}),
		},
		"dst_ip": &proxyReader{
			ProxyReadFunc: ProxyReadFunc(func(name string, proxy http_service.IProxy) (interface{}, bool) {
				return proxy.RemoteIP(), true
			}),
		},
		"dst_port": &proxyReader{
			ProxyReadFunc: ProxyReadFunc(func(name string, proxy http_service.IProxy) (interface{}, bool) {
				return proxy.RemotePort(), true
			}),
		},
		"scheme": &proxyReader{
			ProxyReadFunc: ProxyReadFunc(func(name string, proxy http_service.IProxy) (interface{}, bool) {
				return proxy.URI().Scheme(), true
			}),
			ProxyReadRequestFunc: ProxyReadRequestFunc(func(name string, proxy http_service.IRequest) (interface{}, bool) {
				return proxy.URI().Scheme(), true
			}),
		},
		"method": &proxyReader{
			ProxyReadFunc: ProxyReadFunc(func(name string, proxy http_service.IProxy) (interface{}, bool) {
				return proxy.Method(), true
			}),
			ProxyReadRequestFunc: ProxyReadRequestFunc(func(name string, proxy http_service.IRequest) (interface{}, bool) {
				return proxy.Method(), true
			}),
		},
		"status": &proxyReader{
			ProxyReadFunc: ProxyReadFunc(func(name string, proxy http_service.IProxy) (interface{}, bool) {
				return proxy.StatusCode(), true
			}),
		},
		"path": &proxyReader{
			ProxyReadFunc: ProxyReadFunc(func(name string, proxy http_service.IProxy) (interface{}, bool) {
				return proxy.URI().Path(), true
			}),
			ProxyReadRequestFunc: ProxyReadRequestFunc(func(name string, proxy http_service.IRequest) (interface{}, bool) {
				return proxy.URI().Path(), true
			}),
		},
		"host": &proxyReader{
			ProxyReadFunc: ProxyReadFunc(func(name string, proxy http_service.IProxy) (interface{}, bool) {
				return proxy.Header().Host(), true
			}),
			ProxyReadRequestFunc: ProxyReadRequestFunc(func(name string, proxy http_service.IRequest) (interface{}, bool) {
				return proxy.Header().Host(), true
			}),
		},
		"request_length": &proxyReader{
			ProxyReadFunc: ProxyReadFunc(func(name string, proxy http_service.IProxy) (interface{}, bool) {
				return proxy.ContentLength(), true
			}),
			ProxyReadRequestFunc: ProxyReadRequestFunc(func(name string, proxy http_service.IRequest) (interface{}, bool) {
				return proxy.ContentLength(), true
			}),
		},
		"response_length": &proxyReader{
			ProxyReadFunc: ProxyReadFunc(func(name string, proxy http_service.IProxy) (interface{}, bool) {
				return proxy.ResponseLength(), true
			}),
		},
		"response_body": &proxyReader{
			ProxyReadFunc: ProxyReadFunc(func(name string, proxy http_service.IProxy) (interface{}, bool) {
				body := proxy.ResponseBody()
				encoding := proxy.ResponseHeaders().Get("Content-Encoding")
				if encoding == "gzip" {
					reader, err := gzip.NewReader(bytes.NewReader([]byte(body)))
					if err != nil {
						return "", false
					}
					defer reader.Close()
					data, err := io.ReadAll(reader)
					if err != nil {
						return "", false
					}
					return string(data), true
				}
				return body, true
			}),
		},
		"response_headers": &proxyReader{
			ProxyReadFunc: ProxyReadFunc(func(name string, proxy http_service.IProxy) (interface{}, bool) {
				result := make(map[string]string)
				for key, value := range proxy.ResponseHeaders() {
					result[strings.ToLower(key)] = strings.Join(value, ";")
				}
				return result, true
			}),
		},
		"time": &proxyReader{
			ProxyReadFunc: ProxyReadFunc(func(name string, proxy http_service.IProxy) (interface{}, bool) {
				return proxy.ResponseTime(), true
			}),
		},
		"msec": &proxyReader{
			ProxyReadFunc: ProxyReadFunc(func(name string, proxy http_service.IProxy) (interface{}, bool) {
				return proxy.ProxyTime().UnixMilli(), true
			}),
		},
	}
)

func GetProxyReaders() ProxyReaders {
	return proxyFields
}

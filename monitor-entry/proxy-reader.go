package monitor_entry

import (
	"fmt"
	"strconv"

	http_context "github.com/eolinker/eosc/eocontext/http-context"
	"github.com/eolinker/eosc/log"
)

type ProxyReadFunc func(request http_context.IProxy) (interface{}, bool)

var proxyMetrics = []string{
	"method",
	"host",
	"addr",
	"path",
	"status",
}

var proxyFields = []string{
	"timing",
	"request",
	"response",
}

func ReadProxy(ctx http_context.IHttpContext) []IPoint {
	if len(ctx.Proxies()) < 1 {
		return make([]IPoint, 0, 1)
	}
	labelMetrics := map[string]string{
		"request_id": ctx.RequestId(),
	}
	for key, label := range labels {
		labelMetrics[key] = ctx.GetLabel(label)
	}
	points := make([]IPoint, 0, len(ctx.Proxies()))
	for i, p := range ctx.Proxies() {
		tags := map[string]string{
			"index": strconv.Itoa(i),
		}

		for key, value := range labelMetrics {
			tags[key] = value
		}
		for _, metrics := range proxyMetrics {
			f, has := proxy[metrics]
			if !has {
				log.Error("missing function belong to ", metrics)
				continue
			}
			v, has := f(p)
			if !has {
				continue
			}
			tags[metrics] = v.(string)
		}

		fields := make(map[string]interface{})
		for _, field := range proxyFields {
			f, has := proxy[field]
			if !has {
				log.Error("missing function belong to ", field)
				continue
			}
			v, has := f(p)
			if !has {
				continue
			}
			fields[field] = v
		}
		points = append(points, NewPoint("proxy", tags, fields, ctx.AcceptTime()))
	}
	return points
}

var proxy = map[string]ProxyReadFunc{
	"host": func(proxy http_context.IProxy) (interface{}, bool) {
		return proxy.Header().Host(), true
	},
	"method": func(proxy http_context.IProxy) (interface{}, bool) {
		return proxy.Method(), true
	},
	"path": func(proxy http_context.IProxy) (interface{}, bool) {
		return proxy.URI().Path(), true
	},
	"addr": func(proxy http_context.IProxy) (interface{}, bool) {
		return fmt.Sprintf("%s://%s", proxy.URI().Scheme(), proxy.URI().Host()), true
	},
	"status": func(proxy http_context.IProxy) (interface{}, bool) {
		return proxy.Status(), true
	},
	"timing": func(proxy http_context.IProxy) (interface{}, bool) {
		return proxy.ResponseTime(), true
	},
	"request": func(proxy http_context.IProxy) (interface{}, bool) {
		return proxy.ContentLength(), true
	},
	"response": func(proxy http_context.IProxy) (interface{}, bool) {
		return proxy.ResponseLength(), true
	},
}

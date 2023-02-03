package monitor_entry

import (
	"fmt"

	"github.com/eolinker/eosc/utils"

	http_context "github.com/eolinker/eosc/eocontext/http-context"
	"github.com/eolinker/eosc/log"
)

type ProxyReadFunc func(request http_context.IProxy) (interface{}, bool)

var proxyMetrics = []string{
	"method",
	"host",
	"addr",
}

var proxyFields = []string{
	"timing",
	"request",
	"response",
	"status",
}

func ReadProxy(ctx http_context.IHttpContext) []IPoint {
	if len(ctx.Proxies()) < 1 {
		return make([]IPoint, 0, 1)
	}

	globalLabels := utils.GlobalLabelGet()
	labelMetrics := map[string]string{
		"cluster": globalLabels["cluster_id"],
		"node":    globalLabels["node_id"],
	}
	for key, label := range labels {
		value := ctx.GetLabel(label)
		if value == "" {
			value = "-"
		}
		log.Error("label name: ", key, " label value: ", value)
		labelMetrics[key] = value
	}

	points := make([]IPoint, 0, len(ctx.Proxies()))
	for i, p := range ctx.Proxies() {
		tags := map[string]string{}

		for key, value := range labelMetrics {
			tags[key] = value
		}
		for _, metrics := range proxyMetrics {
			f, has := proxy[metrics]
			if !has {
				log.Error("proxy missing tag function belong to ", metrics)
				continue
			}
			v, has := f(p)
			if !has {
				continue
			}
			tags[metrics] = v.(string)
		}

		fields := map[string]interface{}{
			"index": i,
		}
		for _, field := range proxyFields {
			f, has := proxy[field]
			if !has {
				log.Error("proxy missing field function belong to ", field)
				continue
			}
			v, has := f(p)
			if !has {
				continue
			}
			fields[field] = v
		}
		points = append(points, NewPoint("proxy", tags, fields, p.ProxyTime()))
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
	//"path": func(proxy http_context.IProxy) (interface{}, bool) {
	//	return proxy.URI().Path(), true
	//},
	"addr": func(proxy http_context.IProxy) (interface{}, bool) {
		return fmt.Sprintf("%s://%s", proxy.URI().Scheme(), proxy.URI().Host()), true
	},
	"status": func(proxy http_context.IProxy) (interface{}, bool) {
		return proxy.StatusCode(), true
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

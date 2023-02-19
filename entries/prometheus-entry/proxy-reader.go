package prometheus_entry

import (
	"fmt"
	http_context "github.com/eolinker/eosc/eocontext/http-context"
)

type proxyCollectorReadFunc func(request http_context.IProxy) float64
type proxyLabelReadFunc func(request http_context.IProxy) string

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

var proxyColRead = map[string]proxyCollectorReadFunc{
	"proxy_total": func(proxy http_context.IProxy) float64 {
		return 1
	},
	"proxy_timing": func(proxy http_context.IProxy) float64 {
		return float64(proxy.ResponseTime())
	},
	"proxy_req": func(proxy http_context.IProxy) float64 {
		return float64(proxy.ContentLength())
	},
	"proxy_resp": func(proxy http_context.IProxy) float64 {
		return float64(proxy.ResponseLength())
	},
}

var proxyLabelRead = map[string]proxyLabelReadFunc{
	"host": func(proxy http_context.IProxy) string {
		return proxy.Header().Host()
	},
	"method": func(proxy http_context.IProxy) string {
		return proxy.Method()
	},
	"path": func(proxy http_context.IProxy) string {
		return proxy.URI().Path()
	},
	"addr": func(proxy http_context.IProxy) string {
		return fmt.Sprintf("%s://%s", proxy.URI().Scheme(), proxy.URI().Host())
	},
	"status": func(proxy http_context.IProxy) string {
		return proxy.Status()
	},
}

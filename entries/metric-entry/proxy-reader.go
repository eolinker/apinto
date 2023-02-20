package metric_entry

import (
	"fmt"
	http_context "github.com/eolinker/eosc/eocontext/http-context"
)

type proxyCollectorReadFunc func(request http_context.IProxy) (float64, bool)
type proxyLabelReadFunc func(request http_context.IProxy) string

var proxyColRead = map[string]proxyCollectorReadFunc{
	"proxy_total": func(proxy http_context.IProxy) (float64, bool) {
		return 1, true
	},
	"proxy_timing": func(proxy http_context.IProxy) (float64, bool) {
		return float64(proxy.ResponseTime()), true
	},
	"proxy_req": func(proxy http_context.IProxy) (float64, bool) {
		return float64(proxy.ContentLength()), true
	},
	"proxy_resp": func(proxy http_context.IProxy) (float64, bool) {
		return float64(proxy.ResponseLength()), true
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

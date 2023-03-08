package metric_entry

import (
	http_context "github.com/eolinker/eosc/eocontext/http-context"
)

type proxyCollectorReadFunc func(request http_context.IProxy) (float64, bool)

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

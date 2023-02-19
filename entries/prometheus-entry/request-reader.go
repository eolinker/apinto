package prometheus_entry

import (
	"time"

	http_context "github.com/eolinker/eosc/eocontext/http-context"
)

var requestMetrics = []string{
	"method",
	"host",
}

var requestFields = []string{
	"request",
	"response",
	"retry",
	"timing",
	"status",
}

type reqCollectorReadFunc func(ctx http_context.IHttpContext) float64
type reqLabelReadFunc func(ctx http_context.IHttpContext) string

var reqColRead = map[string]reqCollectorReadFunc{
	"request_total": func(ctx http_context.IHttpContext) float64 {
		return 1
	},
	"request_timing": func(ctx http_context.IHttpContext) float64 {
		return float64(time.Now().Sub(ctx.AcceptTime()).Milliseconds())
	},
	"request_req": func(ctx http_context.IHttpContext) float64 {
		return float64(ctx.Request().ContentLength())
	},
	"request_resp": func(ctx http_context.IHttpContext) float64 {
		return float64(ctx.Response().ContentLength())
	},
	"request_retry": func(ctx http_context.IHttpContext) float64 {
		length := len(ctx.Proxies())
		if length < 1 {
			return 0
		}
		return float64(length - 1)
	},
}

var reqLabelRead = map[string]reqLabelReadFunc{
	"host": func(ctx http_context.IHttpContext) string {
		return ctx.Request().URI().Host()
	},
	"method": func(ctx http_context.IHttpContext) string {
		return ctx.Request().Method()
	},
	"path": func(ctx http_context.IHttpContext) string {
		return ctx.Request().URI().Path()
	},
	"ip": func(ctx http_context.IHttpContext) string {
		return ctx.GetLabel("ip")
	},
	"status": func(ctx http_context.IHttpContext) string {
		return ctx.Response().Status()
	},
}

package metric_entry

import (
	"time"

	http_context "github.com/eolinker/eosc/eocontext/http-context"
)

type reqCollectorReadFunc func(ctx http_context.IHttpContext) (float64, bool)
type reqLabelReadFunc func(ctx http_context.IHttpContext) string

var reqColRead = map[string]reqCollectorReadFunc{
	"request_total": func(ctx http_context.IHttpContext) (float64, bool) {
		return 1, true
	},
	"request_timing": func(ctx http_context.IHttpContext) (float64, bool) {
		return float64(time.Now().Sub(ctx.AcceptTime()).Milliseconds()), true
	},
	"request_req": func(ctx http_context.IHttpContext) (float64, bool) {
		return float64(ctx.Request().ContentLength()), true
	},
	"request_resp": func(ctx http_context.IHttpContext) (float64, bool) {
		return float64(ctx.Response().ContentLength()), true
	},
	"request_retry": func(ctx http_context.IHttpContext) (float64, bool) {
		length := len(ctx.Proxies())
		if length < 1 {
			return 0, true
		}
		return float64(length - 1), true
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
	"status": func(ctx http_context.IHttpContext) string {
		return ctx.Response().Status()
	},
}

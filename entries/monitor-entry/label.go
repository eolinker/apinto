package monitor_entry

import http_context "github.com/eolinker/eosc/eocontext/http-context"

var (
	LabelApi        = "api"
	LabelApp        = "app"
	LabelUpstream   = "upstream"
	LabelHandler    = "handler"
	LabelProvider   = "provider"
	LabelAPIKind    = "api_kind"
	LabelStatusCode = "status_code"
)

var labels = map[string]string{
	LabelApi:        "api",
	LabelApp:        "application",
	LabelHandler:    "handler",
	LabelUpstream:   "service",
	LabelProvider:   "provider",
	LabelAPIKind:    "api_kind",
	LabelStatusCode: "status_code",
}

type GetLabel func(ctx http_context.IHttpContext) string

var readLabel = map[string]GetLabel{
	"status_code": func(ctx http_context.IHttpContext) string {
		statusCode := ctx.Response().StatusCode()
		switch {
		case statusCode >= 200 && statusCode < 300:
			return "2xx"
		case statusCode >= 300 && statusCode < 400:
			return "3xx"
		case statusCode >= 400 && statusCode < 500:
			return "4xx"
		case statusCode >= 500 && statusCode < 600:
			return "5xx"
		default:
			return "other"
		}
	},
}

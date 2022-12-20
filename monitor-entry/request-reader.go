package monitor_entry

import (
	"strings"
	"time"

	"github.com/eolinker/eosc/log"

	http_context "github.com/eolinker/eosc/eocontext/http-context"
)

var requestMetrics = []string{
	"method",
	"host",
	"ip",
	"path",
	"request_id",
	"status",
}

var requestFields = []string{
	"request",
	"response",
	"retry",
	"timing",
}

type RequestReadFunc func(ctx http_context.IHttpContext) (interface{}, bool)

func ReadRequest(ctx http_context.IHttpContext) []IPoint {
	tags := make(map[string]string)
	fields := make(map[string]interface{})
	for _, label := range labels {
		tags[label] = ctx.GetLabel(label)
	}
	for _, metrics := range requestMetrics {
		f, has := request[metrics]
		if !has {
			log.Error("missing function belong to ", metrics)
			continue
		}
		v, has := f(ctx)
		if !has {
			continue
		}
		tags[metrics] = v.(string)
	}

	for _, field := range requestFields {
		f, has := request[field]
		if !has {
			log.Error("missing function belong to ", field)
			continue
		}
		v, has := f(ctx)
		if !has {
			continue
		}
		fields[field] = v
	}

	return []IPoint{NewPoint("request", tags, fields, ctx.AcceptTime())}
}

var request = map[string]RequestReadFunc{
	"host": func(ctx http_context.IHttpContext) (interface{}, bool) {
		return ctx.Request().URI().Host(), true
	},
	"method": func(ctx http_context.IHttpContext) (interface{}, bool) {
		return ctx.Request().Method(), true
	},
	"path": func(ctx http_context.IHttpContext) (interface{}, bool) {
		return ctx.Request().URI().Path(), true
	},
	"ip": func(ctx http_context.IHttpContext) (interface{}, bool) {
		forwardFor := ctx.Request().ForwardIP()
		if forwardFor != "" {
			return strings.Split(forwardFor, ",")[0], true
		}
		realIp := ctx.Request().ReadIP()
		if realIp != "" {
			return realIp, true
		}
		return ctx.Request().RemoteAddr(), true
	},
	"status": func(ctx http_context.IHttpContext) (interface{}, bool) {
		return ctx.Response().Status(), true
	},
	"request_id": func(ctx http_context.IHttpContext) (interface{}, bool) {
		return ctx.RequestId(), true
	},
	"timing": func(ctx http_context.IHttpContext) (interface{}, bool) {
		return time.Now().Sub(ctx.AcceptTime()), true
	},
	"request": func(ctx http_context.IHttpContext) (interface{}, bool) {
		return ctx.Request().ContentLength(), true
	},
	"response": func(ctx http_context.IHttpContext) (interface{}, bool) {
		return ctx.Response().ContentLength(), true
	},
	"retry": func(ctx http_context.IHttpContext) (interface{}, bool) {
		length := len(ctx.Proxies())
		if length < 1 {
			return 0, true
		}
		return length - 1, true
	},
}

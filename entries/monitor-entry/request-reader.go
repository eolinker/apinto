package monitor_entry

import (
	"os"
	"time"

	"github.com/eolinker/eosc/log"

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

type RequestReadFunc func(ctx http_context.IHttpContext) (interface{}, bool)

func ReadRequest(ctx http_context.IHttpContext) []IPoint {
	tags := map[string]string{
		"cluster": os.Getenv("cluster_id"),
		"node":    os.Getenv("node_id"),
	}

	for key, label := range labels {
		value := ctx.GetLabel(label)
		if value == "" {
			value = "-"
		}
		tags[key] = value
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
	fields := make(map[string]interface{})
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
	//"path": func(ctx http_context.IHttpContext) (interface{}, bool) {
	//	return ctx.Request().URI().Path(), true
	//},
	//"ip": func(ctx http_context.IHttpContext) (interface{}, bool) {
	//	return ctx.GetLabel("ip"), true
	//},
	"status": func(ctx http_context.IHttpContext) (interface{}, bool) {
		return ctx.Response().StatusCode(), true
	},
	"timing": func(ctx http_context.IHttpContext) (interface{}, bool) {
		return time.Now().Sub(ctx.AcceptTime()).Milliseconds(), true
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

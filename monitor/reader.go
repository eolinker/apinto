package monitor

import http_context "github.com/eolinker/eosc/eocontext/http-context"

var (
	LabelNode    = "node"
	LabelCluster = "cluster"
	LabelApi     = "api"
	LabelApp     = "app"
	LabelHandler = "handler"
)

var (
	periodRequest = "request"
	periodProxy   = "proxy"
)

type IReader interface {
	Read(period string, ctx http_context.IHttpContext) (interface{}, bool)
}

type ReadFunc func(period string, ctx http_context.IHttpContext) (interface{}, bool)

func (f ReadFunc) Read(period string, ctx http_context.IHttpContext) (interface{}, bool) {
	return f(period, ctx)
}

func ReadRequest(ctx http_context.IHttpContext) IPoint {
	return nil
}

func ReadProxy(ctx http_context.IHttpContext) IPoint {
	return nil
}

func ReadFromValue(label string, ctx http_context.IHttpContext) (string, bool) {
	value := ctx.Value(label)
	if value == nil {
		return "", false
	}
	v, ok := value.(string)
	return v, ok
}

var tags = map[string]IReader{
	"host": ReadFunc(func(period string, ctx http_context.IHttpContext) (interface{}, bool) {
		return ctx.Request().URI().Host(), true
	}),
	"method": ReadFunc(func(period string, ctx http_context.IHttpContext) (interface{}, bool) {
		if period == periodRequest {
			return ctx.Request().Method(), true
		}
		return ctx.Proxy().Method(), true
	}),
	"path": ReadFunc(func(period string, ctx http_context.IHttpContext) (interface{}, bool) {
		if period == periodRequest {
			return ctx.Request().URI().Path(), true
		}
		return ctx.Proxy().URI().Path(), true
	}),
}

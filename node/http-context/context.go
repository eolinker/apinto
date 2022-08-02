package http_context

import (
	"context"
	"fmt"
	"github.com/eolinker/eosc/utils/config"
	"strings"
	"time"

	fasthttp_client "github.com/eolinker/apinto/node/fasthttp-client"

	eoscContext "github.com/eolinker/eosc/eocontext"
	http_service "github.com/eolinker/eosc/eocontext/http-context"
	uuid "github.com/satori/go.uuid"
	"github.com/valyala/fasthttp"
)

var _ http_service.IHttpContext = (*Context)(nil)
var defaultFinisher = new(finishHttp)

//Context fasthttpRequestCtx
type Context struct {
	fastHttpRequestCtx *fasthttp.RequestCtx
	proxyRequest       *ProxyRequest
	proxyRequests      []http_service.IRequest
	requestID          string
	response           *Response

	requestReader *RequestReader
	ctx           context.Context

	completeHandler eoscContext.CompleteHandler

	finishHandler eoscContext.FinishHandler
}

type finishHttp struct {
}

func (f *finishHttp) Finish(ctx eoscContext.EoContext) error {
	target, ok := ctx.(*Context)
	if !ok {
		return nil
	}
	target.finish()
	return nil
}

func (ctx *Context) Complete() eoscContext.CompleteHandler {
	return ctx.completeHandler
}

func (ctx *Context) SetCompleteHandler(handler eoscContext.CompleteHandler) {
	ctx.completeHandler = handler
}

func (ctx *Context) Finish() eoscContext.FinishHandler {
	return ctx.finishHandler
}

func (ctx *Context) SetFinish(handler eoscContext.FinishHandler) {
	ctx.finishHandler = handler
}

func (ctx *Context) Scheme() string {

	return string(ctx.fastHttpRequestCtx.Request.URI().Scheme())
}

func (ctx *Context) Assert(i interface{}) error {
	if v, ok := i.(*http_service.IHttpContext); ok {
		*v = ctx
		return nil
	}
	return fmt.Errorf("not suport:%s", config.TypeNameOf(i))
}

func (ctx *Context) Proxies() []http_service.IRequest {
	return ctx.proxyRequests
}

func (ctx *Context) Response() http_service.IResponse {
	return ctx.response
}

func (ctx *Context) SendTo(address string, timeout time.Duration) error {
	clone := ctx.proxyRequest.clone()
	_, host := readAddress(address)
	clone.URI().SetHost(host)
	ctx.proxyRequests = append(ctx.proxyRequests, clone)
	request := ctx.proxyRequest.Request()

	ctx.response.responseError = fasthttp_client.ProxyTimeout(address, request, &ctx.fastHttpRequestCtx.Response, timeout)

	return ctx.response.responseError

}

func (ctx *Context) Context() context.Context {
	if ctx.ctx == nil {
		ctx.ctx = context.Background()
	}
	return ctx.ctx
}

func (ctx *Context) Value(key interface{}) interface{} {
	return ctx.Context().Value(key)
}

func (ctx *Context) WithValue(key, val interface{}) {
	ctx.ctx = context.WithValue(ctx.Context(), key, val)
}

func (ctx *Context) Proxy() http_service.IRequest {
	return ctx.proxyRequest
}

func (ctx *Context) Request() http_service.IRequestReader {

	return ctx.requestReader
}

//NewContext 创建Context
func NewContext(ctx *fasthttp.RequestCtx) *Context {
	id := uuid.NewV4()
	requestID := id.String()
	newCtx := &Context{
		fastHttpRequestCtx: ctx,
		requestID:          requestID,
		requestReader:      NewRequestReader(&ctx.Request, ctx.RemoteAddr().String()),
		proxyRequest:       NewProxyRequest(&ctx.Request, ctx.RemoteAddr().String()),
		proxyRequests:      make([]http_service.IRequest, 0, 5),
		response:           NewResponse(ctx),
		finishHandler:      defaultFinisher,
	}
	//记录请求时间
	newCtx.WithValue("request_time", ctx.Time())
	return newCtx
}

//RequestId 请求ID
func (ctx *Context) RequestId() string {
	return ctx.requestID
}

//Finish finish
func (ctx *Context) finish() {
	if ctx.response.responseError != nil {
		ctx.fastHttpRequestCtx.SetStatusCode(504)
		ctx.fastHttpRequestCtx.SetBodyString(ctx.response.responseError.Error())
		return
	}

	ctx.requestReader.Finish()
	ctx.proxyRequest.Finish()
	for _, request := range ctx.proxyRequests {
		r, ok := request.(*ProxyRequest)
		if ok {
			r.Finish()
		}
	}
	return
}

func NotFound(ctx *Context) {
	ctx.fastHttpRequestCtx.SetStatusCode(404)
	ctx.fastHttpRequestCtx.SetBody([]byte("404 Not Found"))
}

func readAddress(addr string) (scheme, host string) {
	if i := strings.Index(addr, "://"); i > 0 {
		return strings.ToLower(addr[:i]), addr[i+3:]
	}
	return "http", addr
}

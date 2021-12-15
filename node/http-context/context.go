package http_context

import (
	"context"
	"time"

	fasthttp_client "github.com/eolinker/goku/node/fasthttp-client"

	"github.com/valyala/fasthttp"

	http_service "github.com/eolinker/eosc/http-service"
	uuid "github.com/satori/go.uuid"
)

var _ http_service.IHttpContext = (*Context)(nil)

//Context fasthttpRequestCtx
type Context struct {
	fastHttpRequestCtx *fasthttp.RequestCtx
	proxyRequest       *ProxyRequest
	proxyRequests      []http_service.IRequest
	requestID          string
	response           *Response
	responseError      error
	requestReader      *RequestReader
	ctx                context.Context
}

func (ctx *Context) Proxies() []http_service.IRequest {
	return ctx.proxyRequests
}

//func (ctx *Context) SetField(key, value string) {
//	if ctx.entry == nil {
//		ctx.entry = NewEntry()
//	}
//	ctx.entry.SetField(key, value)
//}
//
//func (ctx *Context) SetChildren(name string, fields []map[string]string) {
//	if ctx.entry == nil {
//		ctx.entry = NewEntry()
//	}
//	ctx.entry.SetChildren(name, fields)
//}

func (ctx *Context) Response() http_service.IResponse {
	return ctx.response
}

func (ctx *Context) ResponseError() error {
	return ctx.responseError
}

type Finish interface {
	Finish() error
}

func (ctx *Context) SendTo(address string, timeout time.Duration) error {
	clone := ctx.proxyRequest.clone()
	clone.URI().SetHost(address)
	ctx.proxyRequests = append(ctx.proxyRequests, clone)
	request := ctx.proxyRequest.Request()

	ctx.responseError = fasthttp_client.ProxyTimeout(address, request, &ctx.fastHttpRequestCtx.Response, timeout)

	return ctx.responseError

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
		requestReader:      NewRequestReader(&ctx.Request, ctx.RemoteIP().String()),
		proxyRequest:       NewProxyRequest(&ctx.Request, ctx.RemoteIP().String()),
		proxyRequests:      make([]http_service.IRequest, 0, 5),
		response:           NewResponse(ctx),
		responseError:      nil,
	}

	return newCtx
}

//RequestId 请求ID
func (ctx *Context) RequestId() string {
	return ctx.requestID
}

//Finish finish
func (ctx *Context) Finish() {
	if ctx.responseError != nil {
		ctx.fastHttpRequestCtx.SetStatusCode(504)
		ctx.fastHttpRequestCtx.SetBodyString(ctx.responseError.Error())
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

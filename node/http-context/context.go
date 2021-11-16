package http_context

import (
	"context"
	"time"

	"github.com/valyala/fasthttp"

	http_service "github.com/eolinker/eosc/http-service"
	uuid "github.com/satori/go.uuid"
)

var _ http_service.IHttpContext = (*Context)(nil)

//Context fasthttpRequestCtx
type Context struct {
	fastHttpRequestCtx *fasthttp.RequestCtx
	requestOrg         *fasthttp.Request
	proxyRequest       *ProxyRequest
	requestID          string
	response           *Response
	responseError      error
	requestReader      *RequestReader
	ctx                context.Context
}

func (ctx *Context) Response() http_service.IResponse {
	return ctx.response
}

func (ctx *Context) ResponseError() error {
	return ctx.responseError
}

func (ctx *Context) SendTo(address string, timeout time.Duration) error {

	return nil

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
		requestOrg:         fasthttp.AcquireRequest(),
		requestID:          requestID,
		requestReader:      NewRequestReader(&ctx.Request, ctx.RemoteIP().String()),
		proxyRequest:       NewProxyRequest(&ctx.Request, ctx.RemoteIP().String()),
		response:           NewResponse(fasthttp.AcquireResponse()),
		responseError:      nil,
	}

	return newCtx
}

//RequestId 请求ID
func (ctx *Context) RequestId() string {
	return ctx.requestID
}

func (ctx *Context) SetResponse(response *fasthttp.Response) {

	ctx.response = NewResponse(response)
	ctx.responseError = nil
}

//Finish finish
func (ctx *Context) Finish() {
	//
	//ctx.proxyResponse.CopyTo(&ctx.fastHttpRequestCtx.Response)
	//
	//fasthttp.ReleaseResponse(ctx.proxyResponse)
	//fasthttp.ReleaseRequest(ctx.proxyRequest.req)
	if ctx.response == nil {
		ctx.fastHttpRequestCtx.SetStatusCode(502)
		ctx.fastHttpRequestCtx.SetBodyString(ctx.responseError.Error())
		return
	}

	ctx.response.WriteTo(ctx.fastHttpRequestCtx)
	ctx.fastHttpRequestCtx.NotModified()
	return
}

func NotFound(ctx *Context) {
	ctx.fastHttpRequestCtx.SetStatusCode(404)
	ctx.fastHttpRequestCtx.SetBody([]byte("404 Not Found"))
}

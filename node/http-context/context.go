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

	proxyRequest *ProxyRequest
	requestID    string
	response     *Response

	requestReader *RequestReader
	ctx           context.Context
}

func (ctx *Context) Response() http_service.IResponse {
	return ctx.response
}

type Finish interface {
	Finish() error
}

func (ctx *Context) SendTo(address string, timeout time.Duration) error {

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
		requestReader:      NewRequestReader(&ctx.Request, ctx.RemoteIP().String()),
		proxyRequest:       NewProxyRequest(&ctx.Request, ctx.RemoteIP().String()),
		response:           NewResponse(ctx),
	}

	return newCtx
}

//RequestId 请求ID
func (ctx *Context) RequestId() string {
	return ctx.requestID
}

//Finish finish
func (ctx *Context) Finish() {
	if ctx.response.responseError != nil {
		ctx.fastHttpRequestCtx.SetStatusCode(504)
		ctx.fastHttpRequestCtx.SetBodyString(ctx.response.responseError.Error())
		return
	}

	ctx.requestReader.Finish()
	ctx.proxyRequest.Finish()
	return
}

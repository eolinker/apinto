package http_context

import (
	"context"
	"encoding/json"

	"github.com/valyala/fasthttp"

	http_service "github.com/eolinker/eosc/http-service"
	uuid "github.com/satori/go.uuid"
)

var _ http_service.IHttpContext = (*Context)(nil)

//Context requestCtx
type Context struct {
	requestCtx *fasthttp.RequestCtx
	requestOrg *fasthttp.Request

	proxyRequest *ProxyRequest

	proxyResponse *fasthttp.Response
	body          []byte
	requestID     string
	RestfulParam  map[string]string
	code          int
	status        string
	response      *Response
	requestReader *RequestReader
	ctx           context.Context
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

func (ctx *Context) Response() http_service.IResponse {
	if ctx.response == nil {
		ctx.response = NewResponse(ctx.proxyResponse)
	}
	return ctx.response
}

func (ctx *Context) Proxy() http_service.IRequest {
	return ctx.proxyRequest
}

func (ctx *Context) SetStatus(code int, status string) {
	ctx.code, ctx.status = code, status
}

func (ctx *Context) Request() http_service.IRequestReader {
	if ctx.requestReader == nil {
		ctx.requestReader = NewRequestReader(ctx.requestOrg, ctx.requestCtx.RemoteAddr().String())
	}
	return ctx.requestReader
}

//NewContext 创建Context
func NewContext(ctx *fasthttp.RequestCtx) *Context {
	id := uuid.NewV4()
	requestID := id.String()
	newRequest := &ctx.Request
	newCtx := &Context{
		requestCtx: ctx,
		requestOrg: fasthttp.AcquireRequest(),
		requestID:  requestID,
	}
	proxyRequest := fasthttp.AcquireRequest()
	newRequest.CopyTo(newCtx.requestOrg)
	newRequest.CopyTo(proxyRequest)

	newCtx.proxyRequest = NewProxyRequest(NewRequestReader(proxyRequest, ""))
	return newCtx
}

//RequestId 请求ID
func (ctx *Context) RequestId() string {
	return ctx.requestID
}

func (ctx *Context) SetBody(body []byte) {
	ctx.requestCtx.SetBody(body)
}

func (ctx *Context) SetResponse(response *fasthttp.Response) {
	ctx.body = response.Body()
	ctx.proxyResponse = response
}

//Finish finish
func (ctx *Context) Finish() {
	ctx.proxyResponse.CopyTo(&ctx.requestCtx.Response)
	return
}

func (ctx *Context) SetError(err error) {
	result := map[string]string{
		"status": "error",
		"msg":    err.Error(),
	}
	errByte, _ := json.Marshal(result)
	ctx.body = errByte
}

func NotFound(ctx *Context) {
	ctx.requestCtx.SetStatusCode(404)
	ctx.requestCtx.SetBody([]byte("404 Not Found"))
}

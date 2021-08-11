package http_context

import (
	"encoding/json"

	"github.com/valyala/fasthttp"

	access_field "github.com/eolinker/goku-eosc/node/common/access-field"
	uuid "github.com/satori/go.uuid"
)

//Context context
type Context struct {
	context       *fasthttp.RequestCtx
	requestOrg    *fasthttp.Request
	proxyRequest  *fasthttp.Request
	proxyResponse *fasthttp.Response
	Body          []byte
	requestID     string
	RestfulParam  map[string]string
	LogFields     *access_field.Fields
	request       IRequest
	labels        map[string]string
	bodyHandler   *BodyRequestHandler
}

//NewContext 创建Context
func NewContext(ctx *fasthttp.RequestCtx) *Context {
	id := uuid.NewV4()
	requestID := id.String()
	newRequest := &ctx.Request
	newCtx := &Context{
		context:      ctx,
		requestOrg:   fasthttp.AcquireRequest(),
		proxyRequest: fasthttp.AcquireRequest(),
		requestID:    requestID,
		LogFields:    access_field.NewFields(),
	}
	newRequest.CopyTo(newCtx.requestOrg)
	newRequest.CopyTo(newCtx.proxyRequest)

	newCtx.LogFields.RequestHeader = newCtx.requestOrg.Header.String()
	newCtx.LogFields.RequestMsg = string(newCtx.Body)
	newCtx.LogFields.RequestMsgSize = len(newCtx.Body)
	newCtx.LogFields.RequestUri = string(newCtx.requestOrg.RequestURI())
	newCtx.LogFields.RequestID = requestID
	return newCtx
}

func (ctx *Context) Labels() map[string]string {
	if ctx.labels == nil {
		ctx.labels = map[string]string{}
	}
	return ctx.labels
}

func (ctx *Context) SetLabels(labels map[string]string) {
	if ctx.labels == nil {
		ctx.labels = make(map[string]string)
	}
	if labels != nil {
		for k, v := range labels {
			ctx.labels[k] = v
		}
	}
}

//RequestId 请求ID
func (ctx *Context) RequestId() string {
	return ctx.requestID
}

func (ctx *Context) Request() IRequest {
	if ctx.request == nil {
		ctx.request = newRequest(ctx.requestOrg)
	}
	return ctx.request
}

func (ctx *Context) RequestOrg() *fasthttp.Request {
	return ctx.requestOrg
}

func (ctx *Context) ProxyRequest() *fasthttp.Request {
	return ctx.proxyRequest
}

func (ctx *Context) ProxyResponse() *fasthttp.Response {
	return ctx.proxyResponse
}

func (ctx *Context) BodyHandler() *BodyRequestHandler {
	if ctx.bodyHandler == nil {
		r := ctx.Request()
		ctx.bodyHandler = newBodyRequestHandler(r.ContentType(), r.RawBody())
	}
	return ctx.bodyHandler
}

func (ctx *Context) SetBody(body []byte) {
	ctx.context.SetBody(body)
}

func (ctx *Context) SetResponse(response *fasthttp.Response) {
	ctx.Body = response.Body()
	ctx.proxyResponse = response
}

//Finish finish
func (ctx *Context) Finish() {
	ctx.LogFields.ResponseMsg = string(ctx.Body)
	ctx.LogFields.ResponseMsgSize = len(ctx.Body)
	ctx.proxyResponse.CopyTo(&ctx.context.Response)
	return
}

func (ctx *Context) SetError(err error) {
	result := map[string]string{
		"status": "error",
		"msg":    err.Error(),
	}
	errByte, _ := json.Marshal(result)
	ctx.Body = errByte
}

func NotFound(ctx *Context) {
	ctx.context.SetStatusCode(404)
	ctx.context.SetBody([]byte("404 Not Found"))
}

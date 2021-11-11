package http_context

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/valyala/fasthttp"

	http_service "github.com/eolinker/eosc/http-service"
	access_field "github.com/eolinker/goku/node/common/access-field"
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

func (ctx *Context) SetStatus(code int, status string) {
	panic("implement me")
}

func (ctx *Context) Request() http_service.RequestReader {
	panic("implement me")
}

func (ctx *Context) ProxyResponse() http_service.ResponseReader {
	panic("implement me")
}

func (ctx *Context) Context() context.Context {
	panic("implement me")
}

func (ctx *Context) Value(key interface{}) interface{} {
	panic("implement me")
}

func (ctx *Context) WithValue(key, val interface{}) {
	panic("implement me")
}

func (ctx *Context) GetHeader(name string) string {
	panic("implement me")
}

func (ctx *Context) Headers() http.Header {
	panic("implement me")
}

func (ctx *Context) SetHeader(key, value string) {
	panic("implement me")
}

func (ctx *Context) AddHeader(key, value string) {
	panic("implement me")
}

func (ctx *Context) DelHeader(key string) {
	panic("implement me")
}

func (ctx *Context) Set() http_service.Header {
	panic("implement me")
}

func (ctx *Context) Append() http_service.Header {
	panic("implement me")
}

func (ctx *Context) Cookie(name string) (*http.Cookie, error) {
	panic("implement me")
}

func (ctx *Context) Cookies() []*http.Cookie {
	panic("implement me")
}

func (ctx *Context) AddCookie(c *http.Cookie) {
	panic("implement me")
}

func (ctx *Context) StatusCode() int {
	panic("implement me")
}

func (ctx *Context) Status() string {
	panic("implement me")
}

func (ctx *Context) GetBody() []byte {
	panic("implement me")
}

func (ctx *Context) Proxy() http_service.Request {
	panic("implement me")
}

func (ctx *Context) SetStoreValue(key string, value interface{}) error {
	panic("implement me")
}

func (ctx *Context) GetStoreValue(key string) (interface{}, bool) {
	panic("implement me")
}

//NewContext 创建Context
func NewContext(ctx *fasthttp.RequestCtx) http_service.IHttpContext {
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

//func (ctx *Context) Request() IRequest {
//	if ctx.request == nil {
//		ctx.request = newRequest(ctx.requestOrg)
//	}
//	return ctx.request
//}

func (ctx *Context) RequestOrg() *fasthttp.Request {
	return ctx.requestOrg
}

func (ctx *Context) ProxyRequest() *fasthttp.Request {
	return ctx.proxyRequest
}

//func (ctx *Context) ProxyResponse() *fasthttp.Response {
//	return ctx.proxyResponse
//}

func (ctx *Context) BodyHandler() *BodyRequestHandler {
	if ctx.bodyHandler == nil {
		r := ctx.Request()
		body, _ := r.RawBody()
		ctx.bodyHandler = newBodyRequestHandler(r.ContentType(), body)
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

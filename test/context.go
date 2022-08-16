package test

import (
	"context"
	"errors"
	"time"

	http_service "github.com/eolinker/eosc/eocontext/http-context"
)

var (
	ErrorNotForm      = errors.New("contentType is not Form")
	ErrorNotMultipart = errors.New("contentType is not Multipart")
	ErrorNotAllowRaw  = errors.New("contentType is not allow Raw")
	ErrorNotSend      = errors.New("not send")
)
var _ http_service.IHttpContext = (*Context)(nil)

type ResponseDO interface {
	Response(request *RequestReader) (*Response, error)
}
type Context struct {
	proxyRequest  *ProxyRequest
	requestID     string
	response      *Response
	responseError error
	requestReader *RequestReader
	ctx           context.Context

	responseHandler ResponseDO
}

func (c *Context) RequestId() string {
	return c.requestID
}

func (c *Context) Context() context.Context {
	if c.ctx == nil {
		c.ctx = context.Background()
	}
	return c.ctx
}

func (c *Context) Value(key interface{}) interface{} {
	return c.Context().Value(key)
}

func (c *Context) WithValue(key, val interface{}) {
	c.ctx = context.WithValue(c.Context(), key, val)
}

func (c *Context) Request() http_service.IRequestReader {
	return c.requestReader
}

func (c *Context) Proxy() http_service.IRequest {
	return c.proxyRequest
}

func (c *Context) Response() http_service.IResponse {
	return c.response
}

func (c *Context) ResponseError() error {
	return c.responseError
}

func (c *Context) SendTo(address string, timeout time.Duration) error {

	if c.responseHandler != nil {
		c.response, c.responseError = c.responseHandler.Response(c.proxyRequest.RequestReader)
	}

	return c.responseError
}

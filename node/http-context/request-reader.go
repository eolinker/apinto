package http_context

import (
	http_service "github.com/eolinker/eosc/http-service"

	"github.com/valyala/fasthttp"
)

var _ http_service.IRequestReader = (*RequestReader)(nil)

type RequestReader struct {
	body    *BodyRequestHandler
	req     *fasthttp.Request
	headers *RequestHeader
	uri     *URIRequest
}

func (r *RequestReader) Header() http_service.IHeaderReader {
	return r.headers
}

func (r *RequestReader) Body() http_service.IBodyDataReader {
	return r.body
}

func (r *RequestReader) URI() http_service.IURIReader {
	return r.uri
}

func (r *RequestReader) ReadIP() string {
	panic("implement me")
}

func (r *RequestReader) ForwardIP() string {
	panic("implement me")
}

func (r *RequestReader) RemoteAddr() string {
	panic("implement me")
}

func NewRequestReader(req *fasthttp.Request) *RequestReader {
	r := &RequestReader{req: req}
	return r
}

func (r *RequestReader) Request() *fasthttp.Request {
	return r.req
}

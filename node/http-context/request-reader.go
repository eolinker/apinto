package http_context

import (
	"strings"

	http_service "github.com/eolinker/eosc/eocontext/http-context"

	"github.com/valyala/fasthttp"
)

var _ http_service.IRequestReader = (*RequestReader)(nil)

type RequestReader struct {
	body       BodyRequestHandler
	req        *fasthttp.Request
	headers    RequestHeader
	uri        URIRequest
	remoteAddr string
	remotePort string
	realIP     string
	length     int
}

func (r *RequestReader) ContentLength() int {
	return r.length
}

func (r *RequestReader) ContentType() string {
	return string(r.req.Header.ContentType())
}

func (r *RequestReader) String() string {
	return r.req.String()
}

func (r *RequestReader) Method() string {
	return string(r.req.Header.Method())
}

func (r *RequestReader) Header() http_service.IHeaderReader {
	return &r.headers
}

func (r *RequestReader) Body() http_service.IBodyDataReader {
	return &r.body
}

func (r *RequestReader) URI() http_service.IURIReader {
	return &r.uri
}

func (r *RequestReader) ReadIP() string {
	if r.realIP == "" {
		realIP := r.headers.GetHeader("x-real-ip")
		if realIP == "" {
			realIP = r.remoteAddr
		}
		r.realIP = realIP
	}
	return r.realIP
}

func (r *RequestReader) ForwardIP() string {
	return r.headers.GetHeader("x-forwarded-for")
}

func (r *RequestReader) RemoteAddr() string {
	return r.remoteAddr
}

func (r *RequestReader) RemotePort() string {
	return r.remotePort
}
func (r *RequestReader) Finish() error {
	r.req = nil
	r.body.reset(nil)
	r.headers.reset(nil)
	r.uri.reset(nil)
	return nil
}
func (r *RequestReader) reset(req *fasthttp.Request, remoteAddr string) {
	r.req = req
	r.remoteAddr = remoteAddr

	r.body.reset(req)

	r.headers.reset(&req.Header)
	r.uri.uri = req.URI()

	idx := strings.LastIndex(remoteAddr, ":")
	if idx != -1 {
		r.remoteAddr = remoteAddr[:idx]
		r.remotePort = remoteAddr[idx+1:]
	}
	length := r.req.Header.ContentLength()
	if length > 0 {
		r.length = length
	}

}

func (r *RequestReader) Request() *fasthttp.Request {
	return r.req
}

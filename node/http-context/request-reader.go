package http_context

import (
	"fmt"
	"strings"

	http_service "github.com/eolinker/eosc/http-service"

	"github.com/valyala/fasthttp"
)

var _ http_service.IRequestReader = (*RequestReader)(nil)

type RequestReader struct {
	body       *BodyRequestHandler
	req        *fasthttp.Request
	headers    *RequestHeader
	uri        *URIRequest
	remoteAddr string
	realIP     string
}

func (r *RequestReader) Finish() error {
	return nil
}

func (r *RequestReader) Method() string {
	return r.headers.Host()
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
	return r.realIP
}

func (r *RequestReader) ForwardIP() string {
	return r.headers.GetHeader("x-forwarded-for")
}

func (r *RequestReader) RemoteAddr() string {
	return r.remoteAddr
}

func NewRequestReader(req *fasthttp.Request, remoteAddr string) *RequestReader {
	r := &RequestReader{
		body:       NewBodyRequestHandler(req),
		req:        req,
		headers:    NewRequestHeader(&req.Header),
		uri:        NewURIRequest(req.URI()),
		remoteAddr: remoteAddr,
	}
	forwardedFor := r.ForwardIP()
	if len(forwardedFor) > 0 {
		if i := strings.Index(forwardedFor, ","); i > 0 {
			r.realIP = forwardedFor[:i]
		} else {
			r.realIP = forwardedFor
		}
		r.headers.SetHeader("x-forwarded-for", fmt.Sprint(forwardedFor, ",", r.remoteAddr))
	} else {
		r.headers.SetHeader("x-forwarded-for", fmt.Sprint(r.remoteAddr))
		r.realIP = remoteAddr
	}
	return r
}

func (r *RequestReader) Request() *fasthttp.Request {
	return r.req
}

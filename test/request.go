package test

import (
	http_service "github.com/eolinker/eosc/eocontext/http-context"
)

type RequestReader struct {
	headers    *RequestHeader
	body       *BodyRequestHandler
	uri        *URIRequest
	remoteAddr string
	method     string
}

func (r *RequestReader) Header() http_service.IHeaderReader {
	return r.headers
}

func (r *RequestReader) Body() http_service.IBodyDataReader {
	return r.body
}

func (r *RequestReader) RemoteAddr() string {
	return r.remoteAddr
}

func (r *RequestReader) ReadIP() string {
	return r.remoteAddr
}

func (r *RequestReader) ForwardIP() string {
	return r.ForwardIP()
}

func (r *RequestReader) URI() http_service.IURIReader {
	return r.uri
}

func (r *RequestReader) Method() string {
	return r.method
}

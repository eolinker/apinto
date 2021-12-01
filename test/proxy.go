package test

import (
	http_service "github.com/eolinker/eosc/http-service"
)

var _ http_service.IRequest = (*ProxyRequest)(nil)

type ProxyRequest struct {
	*RequestReader
}

func (r *ProxyRequest) Finish() error {
	return nil
}
func (r *ProxyRequest) Header() http_service.IHeaderWriter {
	return r.headers
}

func (r *ProxyRequest) Body() http_service.IBodyDataWriter {
	return r.body
}

func (r *ProxyRequest) URI() http_service.IURIWriter {
	return r.uri
}

func (r *ProxyRequest) SetPath(s string) {
	r.URI().SetPath(s)
}

func NewProxyRequest(remoteAddr string) *ProxyRequest {

	return &ProxyRequest{
		RequestReader: &RequestReader{

			remoteAddr: remoteAddr,
		},
	}
}

func (r *ProxyRequest) SetMethod(s string) {
	r.headers.SetMethod(s)
}

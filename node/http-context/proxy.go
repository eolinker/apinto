package http_context

import (
	"github.com/valyala/fasthttp"

	http_service "github.com/eolinker/eosc/http-service"
)

var _ http_service.IRequest = (*ProxyRequest)(nil)

type ProxyRequest struct {
	*RequestReader
}

func (r *ProxyRequest) Header() http_service.IHeaderWriter {
	return r.headers
}

func (r *ProxyRequest) Body() http_service.IBodyDataWriter {
	panic("implement me")
}

func (r *ProxyRequest) URI() http_service.IURIWriter {
	return r.uri
}

func (r *ProxyRequest) SetPath(s string) {
	r.Request().URI().SetPath(s)
}

func NewProxyRequest(request *fasthttp.Request, remoteAddr string) *ProxyRequest {
	proxyRequest := fasthttp.AcquireRequest()
	request.CopyTo(proxyRequest)
	return &ProxyRequest{
		RequestReader: NewRequestReader(proxyRequest, remoteAddr),
	}
}

func (r *ProxyRequest) SetMethod(s string) {
	r.Request().Header.SetMethod(s)
}

package http_context

import (
	"bytes"
	"fmt"

	http_service "github.com/eolinker/eosc/eocontext/http-context"
	"github.com/valyala/fasthttp"
)

var _ http_service.IRequest = (*ProxyRequest)(nil)

type ProxyRequest struct {
	RequestReader
}

//func (r *ProxyRequest) clone() *ProxyRequest {
//	return NewProxyRequest(r.Request(), r.remoteAddr)
//}

func (r *ProxyRequest) Finish() error {
	fasthttp.ReleaseRequest(r.req)
	r.RequestReader.Finish()
	return nil
}
func (r *ProxyRequest) Header() http_service.IHeaderWriter {
	return &r.headers
}

func (r *ProxyRequest) Body() http_service.IBodyDataWriter {
	return &r.body
}

func (r *ProxyRequest) URI() http_service.IURIWriter {
	return &r.uri
}

var (
	xforwardedforKey = []byte("x-forwarded-for")
)

func (r *ProxyRequest) reset(request *fasthttp.Request, remoteAddr string) {
	proxyRequest := fasthttp.AcquireRequest()
	request.CopyTo(proxyRequest)

	forwardedFor := proxyRequest.Header.PeekBytes(xforwardedforKey)
	if len(forwardedFor) > 0 {
		if i := bytes.IndexByte(forwardedFor, ','); i > 0 {
			r.realIP = string(forwardedFor[:i])
		} else {
			r.realIP = string(forwardedFor)
		}
		proxyRequest.Header.Set("x-forwarded-for", fmt.Sprint(string(forwardedFor), ",", r.remoteAddr))
	} else {
		proxyRequest.Header.Set("x-forwarded-for", r.remoteAddr)
		r.realIP = r.remoteAddr
	}

	r.RequestReader.reset(proxyRequest, remoteAddr)
}

//func NewProxyRequest(request *fasthttp.Request, remoteAddr string) *ProxyRequest {
//	proxyRequest := fasthttp.AcquireRequest()
//	request.CopyTo(proxyRequest)
//	return &ProxyRequest{
//		RequestReader: NewRequestReader(proxyRequest, remoteAddr),
//	}
//}

func (r *ProxyRequest) SetMethod(s string) {
	r.Request().Header.SetMethod(s)
}

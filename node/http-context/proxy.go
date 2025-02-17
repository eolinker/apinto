package http_context

import (
	"fmt"

	"github.com/eolinker/eosc/log"

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
	//fasthttp.ReleaseRequest(r.req)
	err := r.RequestReader.Finish()
	if err != nil {
		log.Warn(err)
	}
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

	r.RequestReader.reset(request, remoteAddr)

	forwardedFor := r.req.Header.PeekBytes(xforwardedforKey)
	if r.remoteAddr != "0.0.0.0" {
		if len(forwardedFor) > 0 {
			r.req.Header.Set("x-forwarded-for", fmt.Sprint(string(forwardedFor), ", ", r.remoteAddr))
		} else {
			r.req.Header.Set("x-forwarded-for", r.remoteAddr)
		}
	}
	//if len(forwardedFor) > 0 {
	//	r.req.Header.SetProvider("x-forwarded-for", fmt.Sprint(string(forwardedFor), ", ", r.remoteAddr))
	//} else {
	//	r.req.Header.SetProvider("x-forwarded-for", r.remoteAddr)
	//}
	if r.realIP != "0.0.0.0" {
		r.req.Header.Set("x-real-ip", r.realIP)
	}
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

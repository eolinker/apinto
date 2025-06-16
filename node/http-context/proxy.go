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
	bodyFinishes  []http_service.BodyFinishFunc
	streamHandler []http_service.StreamFunc
}

func (r *ProxyRequest) ProxyBodyFinish(ctx http_service.IHttpContext) {
	for i := len(r.bodyFinishes) - 1; i >= 0; i-- {
		r.bodyFinishes[i](ctx)
	}
}

func (r *ProxyRequest) AppendBodyFinish(fn http_service.BodyFinishFunc) {
	if r.bodyFinishes == nil {
		r.bodyFinishes = make([]http_service.BodyFinishFunc, 0)
	}
	r.bodyFinishes = append(r.bodyFinishes, fn)
}

func (r *ProxyRequest) StreamBodyHandles(ctx http_service.IHttpContext, body []byte) ([]byte, error) {
	tmp := make([]byte, len(body))
	copy(tmp, body)
	var err error
	for _, fn := range r.streamHandler {
		tmp, err = fn(ctx, tmp)
		if err != nil {
			return nil, err
		}
	}
	return tmp, nil
}

func (r *ProxyRequest) AppendStreamBodyHandle(fn http_service.StreamFunc) {
	if r.streamHandler == nil {
		r.streamHandler = make([]http_service.StreamFunc, 0)
	}
	r.streamHandler = append(r.streamHandler, fn)
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
	r.bodyFinishes = nil
	r.streamHandler = nil
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

	if r.realIP != "0.0.0.0" {
		r.req.Header.Set("x-real-ip", r.realIP)
	}
}

func (r *ProxyRequest) SetMethod(s string) {
	r.Request().Header.SetMethod(s)
}

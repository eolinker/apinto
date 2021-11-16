package http_context

import (
	"bytes"
	"net/http"
	"strconv"

	http_service "github.com/eolinker/eosc/http-service"

	"github.com/valyala/fasthttp"
)

var _ http_service.IResponse = (*Response)(nil)
var (
	headerLineSp  = []byte("\r\n")
	headerValueSp = []byte(":")
)

type Response struct {
	*fasthttp.Response
	headersCache http.Header
}

func (r *Response) Headers() http.Header {
	r.initHeader()
	return r.headersCache
}

func NewResponse(response *fasthttp.Response) *Response {
	return &Response{Response: response}
}

func (r *Response) initHeader() {
	if r.headersCache == nil {
		r.headersCache = make(http.Header)

		hs := bytes.Split(r.Response.Header.Header(), headerLineSp)
		for _, h := range hs {
			vs := bytes.Split(h, headerValueSp)
			if len(vs) < 2 {
				if len(vs[0]) == 0 {
					continue
				}
				r.headersCache[string(vs[0])] = []string{""}
				continue
			}
			r.headersCache[string(vs[0])] = []string{string(bytes.TrimSpace(vs[1]))}
		}
	}
}

func (r *Response) GetHeader(name string) string {

	r.initHeader()

	return r.headersCache.Get(name)
}

func (r *Response) GetBody() []byte {
	return r.Response.Body()
}

func (r *Response) StatusCode() int {
	return r.Response.StatusCode()
}

func (r *Response) Status() string {
	return strconv.Itoa(r.Response.StatusCode())
}

func (r *Response) SetHeader(key, value string) {
	if r.headersCache == nil {
		r.Response.Header.Set(key, value)
		return
	}
	r.headersCache.Set(key, value)
}

func (r *Response) AddHeader(key, value string) {
	if r.headersCache != nil {
		r.headersCache.Add(key, value)
	}
	r.Response.Header.Add(key, value)

}

func (r *Response) DelHeader(key string) {
	if r.headersCache != nil {
		r.headersCache.Del(key)
	}
	r.Response.Header.Del(key)

}

func (r *Response) SetStatus(code int, status string) {
	r.Response.SetStatusCode(code)
}

func (r *Response) SetBody(bytes []byte) {
	r.Response.SetBody(bytes)
}

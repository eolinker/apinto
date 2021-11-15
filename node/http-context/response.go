package http_context

import (
	"net/http"
	"strconv"
	"strings"

	http_service "github.com/eolinker/eosc/http-service"

	"github.com/valyala/fasthttp"
)

var _ http_service.IResponse = (*Response)(nil)

type Response struct {
	response *fasthttp.Response
	headers  http.Header
	code     int
	status   string
}

func (r *Response) Headers() http.Header {
	if r.headers == nil {
		r.headers = make(http.Header)
		hs := strings.Split(r.response.Header.String(), "\r\n")
		for _, h := range hs {
			vs := strings.Split(h, ":")
			if len(vs) < 2 {
				if vs[0] == "" {
					continue
				}
				r.headers[vs[0]] = []string{""}
				continue
			}
			r.headers[vs[0]] = []string{strings.TrimSpace(vs[1])}
		}
	}
	return r.headers
}

func NewResponse(response *fasthttp.Response) *Response {
	return &Response{response: response}
}

func (r *Response) initHeader() {
	r.headers = make(http.Header)
	hs := strings.Split(r.response.Header.String(), "\r\n")
	for _, h := range hs {
		vs := strings.Split(h, ":")
		if len(vs) < 2 {
			if vs[0] == "" {
				continue
			}
			r.headers[vs[0]] = []string{""}
			continue
		}
		r.headers[vs[0]] = []string{strings.TrimSpace(vs[1])}
	}
}

func (r *Response) GetHeader(name string) string {
	if r.headers == nil {
		r.initHeader()
	}
	return r.headers.Get(name)
}

func (r *Response) GetBody() []byte {
	return r.response.Body()
}

func (r *Response) StatusCode() int {
	return r.response.StatusCode()
}

func (r *Response) Status() string {
	return strconv.Itoa(r.response.StatusCode())
}

func (r *Response) SetHeader(key, value string) {
	if r.headers == nil {
		r.response.Header.Set(key, value)
		return
	}
	r.headers.Set(key, value)
}

func (r *Response) AddHeader(key, value string) {
	if r.headers == nil {
		r.response.Header.Add(key, value)
		return
	}
	r.headers.Add(key, value)
}

func (r *Response) DelHeader(key string) {
	if r.headers == nil {
		r.response.Header.Del(key)
		return
	}
	r.headers.Del(key)
}

func (r *Response) SetStatus(code int, status string) {
	r.code, r.status = code, status
}

func (r *Response) SetBody(bytes []byte) {
	r.response.SetBody(bytes)
}

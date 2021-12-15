package http_context

import (
	"net/http"
	"strings"

	http_service "github.com/eolinker/eosc/http-service"

	"github.com/valyala/fasthttp"
)

var _ http_service.IHeaderWriter = (*RequestHeader)(nil)

type RequestHeader struct {
	header *fasthttp.RequestHeader
	tmp    http.Header
}

func (h *RequestHeader) RawHeader() string {
	return h.header.String()
}

func NewRequestHeader(header *fasthttp.RequestHeader) *RequestHeader {
	return &RequestHeader{header: header}
}

func (h *RequestHeader) initHeader() {
	if h.tmp == nil {
		h.tmp = make(http.Header)
		hs := strings.Split(h.header.String(), "\r\n")
		for _, t := range hs {
			vs := strings.SplitN(t, ":", 2)
			if len(vs) < 2 {
				if vs[0] == "" {
					continue
				}
				h.tmp[vs[0]] = []string{""}
				continue
			}
			h.tmp[vs[0]] = []string{strings.TrimSpace(vs[1])}
		}
	}
}

func (h *RequestHeader) Host() string {
	return string(h.header.Host())
}

func (h *RequestHeader) GetHeader(name string) string {
	return h.Headers().Get(name)
}

func (h *RequestHeader) Headers() http.Header {
	h.initHeader()
	return h.tmp
}

func (h *RequestHeader) SetHeader(key, value string) {
	if h.tmp != nil {
		h.tmp.Set(key, value)
	}
	h.header.Set(key, value)
}

func (h *RequestHeader) AddHeader(key, value string) {
	if h.tmp != nil {
		h.tmp.Add(key, value)
	}
	h.header.Add(key, value)
}

func (h *RequestHeader) DelHeader(key string) {
	if h.tmp != nil {
		h.tmp.Del(key)
	}
	h.header.Del(key)
}

func (h *RequestHeader) SetHost(host string) {
	if h.tmp != nil {
		h.tmp.Set("Host", host)
	}
	h.header.SetHost(host)
}

type ResponseHeader struct {
	header *fasthttp.ResponseHeader
	tmp    http.Header
}

func NewResponseHeader(header *fasthttp.ResponseHeader) *ResponseHeader {
	return &ResponseHeader{header: header}
}

func (r *ResponseHeader) GetHeader(name string) string {
	return r.Headers().Get(name)
}

func (r *ResponseHeader) Headers() http.Header {

	if r.tmp == nil {
		r.tmp = make(http.Header)
		hs := strings.Split(r.header.String(), "\r\n")
		for _, t := range hs {
			vs := strings.Split(t, ":")
			if len(vs) < 2 {
				if vs[0] == "" {
					continue
				}
				r.tmp[vs[0]] = []string{""}
				continue
			}
			r.tmp[vs[0]] = []string{strings.TrimSpace(vs[1])}
		}
	}
	return r.tmp
}

func (r *ResponseHeader) SetHeader(key, value string) {
	if r.tmp != nil {
		r.tmp.Set(key, value)
	}
	r.header.Set(key, value)
}

func (r *ResponseHeader) AddHeader(key, value string) {
	if r.tmp != nil {
		r.tmp.Add(key, value)
	}
	r.header.Add(key, value)
}

func (r *ResponseHeader) DelHeader(key string) {
	if r.tmp != nil {
		r.tmp.Del(key)
	}
	r.header.Del(key)
}

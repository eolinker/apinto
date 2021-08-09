package http_context

import (
	"strings"

	"github.com/valyala/fasthttp"
)

type Value map[string]string

func (h Value) Get(key string) (string, bool) {
	v, ok := h[key]
	return v, ok
}

type IRequest interface {
	Host() string
	Method() string
	Path() string
	ContentType() string
	Header() Value
	Query() Value
	RawQuery() string
	RawBody() []byte
}

type Request struct {
	req         *fasthttp.Request
	path        string
	host        string
	method      string
	header      Value
	query       Value
	rawQuery    string
	rawBody     []byte
	contentType string
}

func (r *Request) Host() string {
	if r.host == "" {
		r.host = strings.Split(string(r.req.Header.Host()), ":")[0]
	}
	return r.host
}

func (r *Request) Method() string {
	if r.method == "" {
		r.method = string(r.req.Header.Method())
	}
	return r.method
}

func (r *Request) Path() string {
	if r.path == "" {
		r.path = string(r.req.URI().Path())
	}
	return r.path
}

func (r *Request) Header() Value {
	if r.header == nil {
		r.header = make(Value)
		hs := strings.Split(r.req.Header.String(), "\r\n")
		for _, h := range hs {
			vs := strings.Split(h, ":")
			if len(vs) < 2 {
				if vs[0] == "" {
					continue
				}
				r.header[vs[0]] = ""
				continue
			}
			r.header[vs[0]] = strings.TrimSpace(vs[1])

		}
	}
	return r.header
}

func (r *Request) Query() Value {
	if r.rawQuery == "" {
		r.rawQuery = string(r.req.URI().QueryString())
	}
	if r.query == nil {
		qs := strings.Split(r.rawQuery, "&")
		for _, q := range qs {
			vs := strings.Split(q, ":")
			if len(vs) < 2 {
				if vs[0] == "" {
					continue
				}
				r.query[vs[0]] = ""
				continue
			}
			r.query[vs[0]] = strings.TrimSpace(vs[1])
		}
	}
	return r.query
}

func (r *Request) RawQuery() string {
	if r.rawQuery == "" {
		r.rawQuery = string(r.req.URI().QueryString())
	}
	return r.rawQuery
}

func (r *Request) RawBody() []byte {
	if r.rawBody == nil {
		r.rawBody = r.req.Body()
	}
	return r.rawBody
}

func (r *Request) ContentType() string {
	if r.contentType == "" {
		r.contentType = string(r.req.Header.ContentType())
	}
	return r.contentType
}

func newRequest(req *fasthttp.Request) IRequest {
	req.Header.ContentType()
	newReq := &Request{
		req: req,
	}
	return newReq
}

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
	return r.host
}

func (r *Request) Method() string {
	return r.method
}

func (r *Request) Path() string {
	return r.path
}

func (r *Request) Header() Value {
	return r.header
}

func (r *Request) Query() Value {
	return r.query
}

func (r *Request) RawQuery() string {
	return r.rawQuery
}

func (r *Request) RawBody() []byte {
	return r.rawBody
}

func (r *Request) ContentType() string {
	return r.contentType
}

func NewRequest(req fasthttp.Request) *Request {
	newReq := &Request{
		path:     string(req.URI().Path()),
		host:     strings.Split(string(req.Header.Host()), ":")[0],
		method:   string(req.Header.Method()),
		header:   Value{},
		query:    Value{},
		rawQuery: string(req.URI().QueryString()),
		rawBody:  req.Body(),
	}

	hs := strings.Split(req.Header.String(), "\r\n")
	for _, h := range hs {
		vs := strings.Split(h, ":")
		if len(vs) < 2 {
			if vs[0] == "" {
				continue
			}
			newReq.header[vs[0]] = ""
			continue
		}
		newReq.header[vs[0]] = strings.TrimSpace(vs[1])
		if vs[0] == "Content-Type" {
			newReq.contentType = vs[1]
		}
	}
	qs := strings.Split(newReq.rawQuery, "&")
	for _, q := range qs {
		vs := strings.Split(q, ":")
		if len(vs) < 2 {
			if vs[0] == "" {
				continue
			}
			newReq.query[vs[0]] = ""
			continue
		}
		newReq.query[vs[0]] = strings.TrimSpace(vs[1])
	}
	return newReq
}

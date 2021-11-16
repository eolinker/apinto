package http_context

import (
	"net/url"
	"strings"

	http_service "github.com/eolinker/eosc/http-service"
	"github.com/valyala/fasthttp"
)

var _ http_service.IURIWriter = (*URIRequest)(nil)

type URIRequest struct {
	uri   *fasthttp.URI
	query *url.Values
}

func (U *URIRequest) initQuery() {

	if U.query == nil {
		U.query = make(url.Values)
		qs := strings.Split(r.rawQuery, "&")
		for _, q := range qs {
			vs := strings.Split(q, "=")
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

func NewURIRequest(uri *fasthttp.URI) *URIRequest {
	return &URIRequest{uri: uri}
}

func (U *URIRequest) RequestURI() string {
	return string(U.uri.RequestURI())
}

func (U *URIRequest) Scheme() string {
	return string(U.uri.Scheme())
}

func (U *URIRequest) RawURL() string {
	return string(U.uri.FullURI())
}

func (U *URIRequest) GetQuery(key string) string {
	panic("implement me")
}

func (U *URIRequest) RawQuery() string {
	panic("implement me")
}

func (U *URIRequest) SetMethod(method string) {
	panic("implement me")
}

func (U *URIRequest) SetRequestURI(uri string) {
	panic("implement me")
}

func (U *URIRequest) SetPath(s string) {
	panic("implement me")
}

func (U *URIRequest) SetHost(host string) {
	panic("implement me")
}

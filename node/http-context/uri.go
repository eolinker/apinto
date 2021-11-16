package http_context

import (
	http_service "github.com/eolinker/eosc/http-service"
	"github.com/valyala/fasthttp"
)

var _ http_service.IURIWriter = (*URIRequest)(nil)

type URIRequest struct {
	uri *fasthttp.URI
}

func (ur *URIRequest) Host() string {
	return string(ur.uri.Host())
}

func (ur *URIRequest) SetQuery(key, value string) {

	ur.uri.QueryArgs().Set(key, value)
}

func (ur *URIRequest) AddQuery(key, value string) {
	ur.uri.QueryArgs().Add(key, value)
}

func (ur *URIRequest) DelQuery(key string) {
	ur.uri.QueryArgs().Del(key)
}

func (ur *URIRequest) SetRawQuery(raw string) {
	ur.uri.SetQueryString(raw)
}

func NewURIRequest(uri *fasthttp.URI) *URIRequest {
	return &URIRequest{uri: uri}
}

func (ur *URIRequest) RequestURI() string {
	return string(ur.uri.RequestURI())
}

func (ur *URIRequest) Scheme() string {
	return string(ur.uri.Scheme())
}

func (ur *URIRequest) RawURL() string {
	return string(ur.uri.FullURI())
}

func (ur *URIRequest) GetQuery(key string) string {

	return string(ur.uri.QueryArgs().Peek(key))
}

func (ur *URIRequest) RawQuery() string {
	return string(ur.uri.QueryString())
}

func (ur *URIRequest) SetPath(s string) {
	ur.uri.SetPath(s)

}

func (ur *URIRequest) SetHost(host string) {
	ur.uri.SetHost(host)
}

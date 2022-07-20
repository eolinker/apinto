package test

import (
	"net/url"

	http_service "github.com/eolinker/eosc/context/http-context"
)

var _ http_service.IURIWriter = (*URIRequest)(nil)

type URIRequest struct {
	url *url.URL
}

func (u *URIRequest) RequestURI() string {
	return u.url.RequestURI()
}

func (u *URIRequest) Scheme() string {
	return u.url.Scheme
}

func (u *URIRequest) RawURL() string {
	return u.RawURL()
}

func (u *URIRequest) Host() string {
	return u.url.Host
}

func (u *URIRequest) Path() string {
	return u.url.Path
}

func (u *URIRequest) GetQuery(key string) string {
	return u.url.Query().Get(key)
}

func (u *URIRequest) RawQuery() string {
	return u.url.RawQuery
}

func (u *URIRequest) SetQuery(key, value string) {
	query := u.url.Query()
	query.Set(key, value)
	u.url.RawQuery = query.Encode()
}

func (u *URIRequest) AddQuery(key, value string) {
	query := u.url.Query()
	query.Add(key, value)
	u.url.RawQuery = query.Encode()
}

func (u *URIRequest) DelQuery(key string) {
	query := u.url.Query()
	query.Del(key)
	u.url.RawQuery = query.Encode()
}

func (u *URIRequest) SetRawQuery(raw string) {
	u.url.RawQuery = raw
}

func (u *URIRequest) SetPath(s string) {
	u.url.Path = s
}

func (u *URIRequest) SetScheme(scheme string) {
	u.url.Scheme = scheme
}

func (u *URIRequest) SetHost(host string) {
	u.url.Host = host
}

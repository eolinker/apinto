package http_context

import http_service "github.com/eolinker/eosc/eocontext/http-context"

type requestAgent struct {
	http_service.IRequest
	host      string
	scheme    string
	hostAgent *UrlAgent
}

func newRequestAgent(IRequest http_service.IRequest, host string, scheme string) *requestAgent {
	return &requestAgent{IRequest: IRequest, host: host, scheme: scheme}
}
func (a *requestAgent) URI() http_service.IURIWriter {
	if a.hostAgent == nil {
		a.hostAgent = NewUrlAgent(a.IRequest.URI(), a.host, a.scheme)
	}
	return a.hostAgent
}

type UrlAgent struct {
	http_service.IURIWriter
	host   string
	scheme string
}

func (u *UrlAgent) SetScheme(scheme string) {
	u.scheme = scheme
}
func (u *UrlAgent) Scheme() string {
	return u.scheme
}

func (u *UrlAgent) Host() string {
	return u.host
}

func (u *UrlAgent) SetHost(host string) {
	u.host = host
}

func NewUrlAgent(IURIWriter http_service.IURIWriter, host string, scheme string) *UrlAgent {
	return &UrlAgent{IURIWriter: IURIWriter, host: host, scheme: scheme}
}

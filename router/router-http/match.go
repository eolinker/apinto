package router_http

import (
	http_context "github.com/eolinker/goku-eosc/node/http-context"

	"github.com/eolinker/goku-eosc/router"
	"github.com/eolinker/goku-eosc/service"
)

type IMatcher interface {
	Match(req http_context.IRequest) (service.IService, router.IEndPoint, bool)
}

type Matcher struct {
	r        router.IRouter
	services map[string]service.IService
}

func (m *Matcher) Match(req http_context.IRequest) (service.IService, router.IEndPoint, bool) {

	sources := newHttpSources(req)
	endpoint, has := m.r.Router(sources)
	if !has {
		return nil, nil, false
	}

	s, has := m.services[endpoint.Target()]

	return s, endpoint, has
}

type HttpSources struct {
	req http_context.IRequest
}

func newHttpSources(req http_context.IRequest) *HttpSources {
	return &HttpSources{req: req}
}

func (h *HttpSources) Get(cmd string) (string, bool) {
	if isHost(cmd) {
		return h.req.Host(), true
	}
	if isMethod(cmd) {
		return h.req.Method(), true
	}

	if isLocation(cmd) {
		return h.req.Path(), true
	}
	if hn, yes := headerName(cmd); yes {
		if vs, has := h.req.Header().Get(hn); has {
			if len(vs) == 0 {
				return "", true
			}
			return vs, true
		}
	}

	if qn, yes := queryName(cmd); yes {
		if vs, has := h.req.Query().Get(qn); has {
			if len(vs) == 0 {
				return "", true
			}
			return vs, true
		}
	}
	return "", false
}

package router_http

import (
	"net/http"
	"strings"

	"github.com/eolinker/goku-eosc/router"
	"github.com/eolinker/goku-eosc/service"
)

type IMatcher interface {
	Match(req *http.Request) (service.IService, router.IEndPoint, bool)
}

type Matcher struct {
	r        router.IRouter
	services map[string]service.IService
}

func (m *Matcher) Match(req *http.Request) (service.IService, router.IEndPoint, bool) {

	sources := newHttpSources(req)
	endpoint, has := m.r.Router(sources)
	if !has {
		return nil, nil, false
	}

	s, has := m.services[endpoint.Target()]

	return s, endpoint, has
}

type HttpSources struct {
	req *http.Request
}

func newHttpSources(req *http.Request) *HttpSources {
	index := strings.Index(req.Host, ":")
	if index > 0 {
		req.Host = req.Host[:index]
	}
	return &HttpSources{req: req}
}

func (h *HttpSources) Get(cmd string) (string, bool) {

	if isHost(cmd) {
		return h.req.Host, true
	}
	if isMethod(cmd){
		return h.req.Method,true
	}

	if isLocation(cmd) {
		return h.req.URL.Path, true
	}
	if hn, yes := headerName(cmd); yes {
		if vs, has := h.req.Header[hn]; has {
			if len(vs) == 0 {
				return "", true
			}
			return vs[0], true
		}
	}

	if qn, yes := queryName(cmd); yes {
		if vs, has := h.req.URL.Query()[qn]; has {
			if len(vs) == 0 {
				return "", true
			}
			return vs[0], true
		}
	}
	return "", false
}

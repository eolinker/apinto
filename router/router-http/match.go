package router_http

import (
	"strings"

	"github.com/valyala/fasthttp"

	"github.com/eolinker/goku-eosc/router"
	"github.com/eolinker/goku-eosc/service"
)

type IMatcher interface {
	Match(req *fasthttp.Request) (service.IService, router.IEndPoint, bool)
}

type Matcher struct {
	r        router.IRouter
	services map[string]service.IService
}

func (m *Matcher) Match(req *fasthttp.Request) (service.IService, router.IEndPoint, bool) {

	sources := newHttpSources(req)
	endpoint, has := m.r.Router(sources)
	if !has {
		return nil, nil, false
	}

	s, has := m.services[endpoint.Target()]

	return s, endpoint, has
}

type HttpSources struct {
	method   string
	host     string
	location string
	header   map[string]string
	queries  map[string]string
}

func newHttpSources(req *fasthttp.Request) *HttpSources {
	sources := &HttpSources{
		method:   string(req.Header.Method()),
		host:     string(req.Header.Host()),
		location: string(req.URI().Path()),
		header:   make(map[string]string),
		queries:  make(map[string]string),
	}
	sources.host = string(req.Host())
	index := strings.Index(sources.host, ":")
	if index > 0 {
		sources.host = sources.host[:index]
	}
	hs := strings.Split(req.Header.String(), "\r\n")
	for _, h := range hs {
		vs := strings.Split(h, ":")
		if len(vs) < 2 {
			if vs[0] == "" {
				continue
			}
			sources.header[vs[0]] = ""
			continue
		}
		sources.header[vs[0]] = strings.TrimSpace(vs[1])
	}
	qs := strings.Split(string(req.URI().QueryString()), "&")
	for _, q := range qs {
		vs := strings.Split(q, ":")
		if len(vs) < 2 {
			if vs[0] == "" {
				continue
			}
			sources.queries[vs[0]] = ""
			continue
		}
		sources.queries[vs[0]] = strings.TrimSpace(vs[1])
	}
	return sources
}

func (h *HttpSources) Get(cmd string) (string, bool) {
	if isHost(cmd) {
		return h.host, true
	}
	if isMethod(cmd) {
		return h.method, true
	}

	if isLocation(cmd) {
		return h.location, true
	}
	if hn, yes := headerName(cmd); yes {
		if vs, has := h.header[hn]; has {
			if len(vs) == 0 {
				return "", true
			}
			return vs, true
		}
	}

	if qn, yes := queryName(cmd); yes {
		if vs, has := h.queries[qn]; has {
			if len(vs) == 0 {
				return "", true
			}
			return vs, true
		}
	}
	return "", false
}

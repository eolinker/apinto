package grpc_router

import (
	"errors"
	"fmt"
	"sort"
	"strconv"

	"github.com/eolinker/apinto/checker"
	"github.com/eolinker/apinto/router"
)

var ErrorDuplicate = errors.New("duplicate")

type Root struct {
	ports map[int]*Ports
}

func (r *Root) Build() router.IMatcher {

	portsHandlers := make(map[string]router.IMatcher)
	for p, c := range r.ports {
		name := strconv.Itoa(p)
		if p == 0 {
			name = router.All
		}
		portsHandlers[name] = c.Build()
	}
	return newPortMatcher(portsHandlers)
}

type Ports struct {
	paths map[string]*Paths
}

func (p *Ports) Build() router.IMatcher {
	pathMatcher := make(map[string]router.IMatcher)
	for m, c := range p.paths {
		pathMatcher[m] = c.Build()
	}
	return newPathMatcher(pathMatcher)
}

//type Hosts struct {
//	paths map[string]*Paths
//}
//
//func (h *Hosts) Build() router.IMatcher {
//	pathMatcher := make(map[string]router.IMatcher)
//	for m, c := range h.paths {
//		pathMatcher[m] = c.Build()
//	}
//	return newPathMatcher(pathMatcher)
//}

type Paths struct {
	handlers map[string]*Handler
	checker  checker.Checker
}

func (p *Paths) Build() router.IMatcher {
	if len(p.handlers) == 0 {
		return &EmptyMatcher{handler: nil, has: false}
	}

	if all, has := p.handlers[router.All]; has {
		if len(p.handlers) == 1 {
			return &EmptyMatcher{handler: all.handler, has: true}
		}
	}

	nexts := make(AppendMatchers, 0, len(p.handlers))
	for _, h := range p.handlers {
		nexts = append(nexts, &AppendMatcher{
			handler:  h.handler,
			checkers: Parse(h.rules),
		})
	}
	sort.Sort(nexts)
	return nexts
}

type Handler struct {
	id      string
	handler router.IRouterHandler
	rules   []router.AppendRule
}

func (h *Handler) Build() router.IMatcher {
	return &AppendMatcher{
		handler:  h.handler,
		checkers: Parse(h.rules),
	}
}

func NewRoot() *Root {
	return &Root{
		ports: map[int]*Ports{},
	}
}

func NewPorts() *Ports {
	return &Ports{
		paths: map[string]*Paths{},
	}
}

func NewPaths(checker checker.Checker) *Paths {
	return &Paths{
		checker:  checker,
		handlers: map[string]*Handler{},
	}
}

func NewHandler(id string, handler router.IRouterHandler, appends []router.AppendRule) *Handler {
	return &Handler{id: id, handler: handler, rules: appends}
}
func (r *Root) Add(id string, handler router.IRouterHandler, port int, service string, method string, append []router.AppendRule) error {
	if r.ports == nil {
		r.ports = make(map[int]*Ports)
	}
	pN, has := r.ports[port]
	if !has {
		pN = NewPorts()
		r.ports[port] = pN
	}
	err := pN.Add(id, handler, service, method, append)
	if err != nil {
		return fmt.Errorf("port=%d %w", port, err)
	}
	return nil
}

func (p *Ports) Add(id string, handler router.IRouterHandler, service string, method string, append []router.AppendRule) error {
	err := p.add(id, handler, service, method, append)
	if err != nil {
		return err
	}
	return nil
}
func (p *Ports) add(id string, handler router.IRouterHandler, service string, method string, append []router.AppendRule) error {
	path := fmt.Sprintf("/%s/%s", service, method)
	ck, err := checker.Parse(path)
	if err != nil {
		return fmt.Errorf("path=%s %w", path, err)
	}
	v, has := p.paths[path]
	if !has {
		v = NewPaths(ck)
		p.paths[path] = v
	}

	err = v.Add(id, handler, append)
	if err != nil {
		return fmt.Errorf("path=%s %w", path, err)
	}
	return nil
}

func (p *Paths) Add(id string, handler router.IRouterHandler, append []router.AppendRule) error {

	key := router.Key(append)
	h, has := p.handlers[key]
	if has && h.id != id {
		return fmt.Errorf(" append{%s}:%w for (%s %s) ", key, ErrorDuplicate, h.id, id)
	}
	p.handlers[key] = NewHandler(id, handler, append)
	return nil
}

type IBuilder interface {
	Build() router.IMatcher
}

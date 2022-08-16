package http_router

import (
	"errors"
	"fmt"
	"github.com/eolinker/apinto/checker"
	"sort"
	"strconv"
)

var ErrorDuplicate = errors.New("duplicate")

type Root struct {
	ports map[int]*Ports
}

func (r *Root) Build() IMatcher {

	portsHandlers := make(map[string]IMatcher)
	for p, c := range r.ports {
		name := strconv.Itoa(p)
		if p == 0 {
			name = All
		}
		portsHandlers[name] = c.Build()
	}
	return newPortMatcher(portsHandlers)
}

type Ports struct {
	hosts map[string]*Hosts
}

func (p *Ports) Build() IMatcher {
	hostMatchers := make(map[string]IMatcher)
	for h, c := range p.hosts {
		hostMatchers[h] = c.Build()
	}
	return newHostMatcher(hostMatchers)
}

type Hosts struct {
	methods map[string]*Methods
}

func (h *Hosts) Build() IMatcher {
	methodMatchers := make(map[string]IMatcher)
	for m, c := range h.methods {
		methodMatchers[m] = c.Build()
	}
	return newMethodMatcher(methodMatchers, nil)
}

type Methods struct {
	paths map[string]*Paths
}

func (m *Methods) Build() IMatcher {

	checkers := make([]*CheckerHandler, 0, len(m.paths))
	equals := make(map[string]IMatcher, len(m.paths))
	for _, next := range m.paths {
		if next.checker.CheckType() == checker.CheckTypeEqual {
			equals[next.checker.Value()] = next.Build()
			continue
		}

		checkers = append(checkers, &CheckerHandler{
			checker: next.checker,
			next:    next.Build(),
		})
	}
	return NewPathMatcher(equals, checkers)
}

type Paths struct {
	handlers map[string]*Handler
	checker  checker.Checker
}

func (p *Paths) Build() IMatcher {
	if len(p.handlers) == 0 {
		return &EmptyMatcher{handler: nil, has: false}
	}

	if all, has := p.handlers[All]; has {
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
	handler IRouterHandler
	rules   []AppendRule
}

func (h *Handler) Build() IMatcher {
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
		hosts: map[string]*Hosts{},
	}
}
func NewHosts() *Hosts {
	return &Hosts{
		methods: map[string]*Methods{},
	}
}
func NewMethods() *Methods {
	return &Methods{paths: map[string]*Paths{}}
}
func NewPaths(checker checker.Checker) *Paths {
	return &Paths{
		checker:  checker,
		handlers: map[string]*Handler{},
	}
}

func NewHandler(id string, handler IRouterHandler, appends []AppendRule) *Handler {
	return &Handler{id: id, handler: handler, rules: appends}
}
func (r *Root) Add(id string, handler IRouterHandler, port int, hosts []string, methods []string, path string, append []AppendRule) error {
	if r.ports == nil {
		r.ports = make(map[int]*Ports)
	}
	pN, has := r.ports[port]
	if !has {
		pN = NewPorts()
		r.ports[port] = pN
	}
	err := pN.Add(id, handler, hosts, methods, path, append)
	if err != nil {
		return fmt.Errorf("port=%d %w", port, err)
	}
	return nil
}

func (p *Ports) Add(id string, handler IRouterHandler, hosts []string, methods []string, path string, append []AppendRule) error {

	if len(hosts) == 0 {
		return p.add(id, handler, All, methods, path, append)
	}
	for _, host := range hosts {
		err := p.add(id, handler, host, methods, path, append)
		if err != nil {
			return err
		}
	}
	return nil
}
func (p *Ports) add(id string, handler IRouterHandler, host string, methods []string, path string, append []AppendRule) error {
	hN, has := p.hosts[host]
	if !has {
		hN = NewHosts()
		p.hosts[host] = hN
	}
	err := hN.Add(id, handler, methods, path, append)
	if err != nil {
		return fmt.Errorf("host=%s %w", host, err)
	}
	return nil
}

func (h *Hosts) add(id string, handler IRouterHandler, method string, path string, append []AppendRule) error {
	methods, has := h.methods[method]
	if !has {
		methods = NewMethods()
		h.methods[method] = methods
	}
	err := methods.Add(id, handler, path, append)
	if err != nil {
		return fmt.Errorf("method=%s %w", method, err)
	}
	return nil
}
func (h *Hosts) Add(id string, handler IRouterHandler, methods []string, path string, append []AppendRule) error {
	if len(methods) == 0 {
		return h.add(id, handler, All, path, append)
	}
	for _, m := range methods {
		err := h.add(id, handler, m, path, append)
		if err != nil {
			return err
		}
	}
	return nil
}

func (m *Methods) Add(id string, handler IRouterHandler, path string, append []AppendRule) error {
	ck, err := checker.Parse(path)
	if err != nil {
		return fmt.Errorf("path=%s %w", path, err)
	}
	key := ck.Key()
	p, has := m.paths[key]
	if !has {
		p = NewPaths(ck)
		m.paths[key] = p
	}

	err = p.Add(id, handler, append)
	if err != nil {
		return fmt.Errorf("path=%s %w", key, err)
	}
	return nil
}

func (p *Paths) Add(id string, handler IRouterHandler, append []AppendRule) error {

	key := Key(append)
	h, has := p.handlers[key]
	if has && h.id != id {
		return fmt.Errorf(" append{%s}:%w for (%s %s) ", key, ErrorDuplicate, h.id, id)
	}
	p.handlers[key] = NewHandler(id, handler, append)
	return nil
}

type IBuilder interface {
	Build() IMatcher
}

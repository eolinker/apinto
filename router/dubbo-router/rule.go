package dubbo_router

import (
	"errors"
	"fmt"
	"github.com/eolinker/apinto/checker"
	"github.com/eolinker/apinto/router"
	"sort"
	"strconv"
)

var ErrorDuplicate = errors.New("duplicate")

type Root struct {
	ports map[int]*Ports
}

type Ports struct {
	hosts map[string]*Hosts
}
type Hosts struct {
	paths map[string]*Paths
}

type Paths struct {
	handlers map[string]*Handler
	checker  checker.Checker
}

type Handler struct {
	id      string
	handler router.IRouterHandler
	rules   []router.AppendRule
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

func (p *Ports) Build() router.IMatcher {
	hostMatchers := make(map[string]router.IMatcher)
	for h, c := range p.hosts {
		hostMatchers[h] = c.Build()
	}
	return newHostMatcher(hostMatchers)
}

func (h *Hosts) Build() router.IMatcher {
	pathMatcher := make(map[string]router.IMatcher)
	for m, c := range h.paths {
		pathMatcher[m] = c.Build()
	}
	return newPathMatcher(pathMatcher)
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

func (r *Root) Add(id string, handler router.IRouterHandler, port int, hosts []string, service string, method string, append []router.AppendRule) error {
	if r.ports == nil {
		r.ports = make(map[int]*Ports)
	}
	pN, has := r.ports[port]
	if !has {
		pN = NewPorts()
		r.ports[port] = pN
	}
	err := pN.Add(id, handler, hosts, service, method, append)
	if err != nil {
		return fmt.Errorf("port=%d %w", port, err)
	}
	return nil
}

func (p *Ports) Add(id string, handler router.IRouterHandler, hosts []string, service string, method string, append []router.AppendRule) error {

	if len(hosts) == 0 {
		return p.add(id, handler, router.All, service, method, append)
	}
	for _, host := range hosts {
		err := p.add(id, handler, host, service, method, append)
		if err != nil {
			return err
		}
	}
	return nil
}
func (p *Ports) add(id string, handler router.IRouterHandler, host string, services string, method string, append []router.AppendRule) error {
	hN, has := p.hosts[host]
	if !has {
		hN = NewHosts()
		p.hosts[host] = hN
	}
	err := hN.Add(id, handler, services, method, append)
	if err != nil {
		return fmt.Errorf("host=%s %w", host, err)
	}
	return nil
}

func (h *Hosts) add(id string, handler router.IRouterHandler, service string, method string, append []router.AppendRule) error {
	path := fmt.Sprintf("%s/%s", service, method)
	ck, err := checker.Parse(path)
	if err != nil {
		return fmt.Errorf("path=%s %w", path, err)
	}
	p, has := h.paths[path]
	if !has {
		p = NewPaths(ck)
		h.paths[path] = p
	}

	err = p.Add(id, handler, append)
	if err != nil {
		return fmt.Errorf("path=%s %w", path, err)
	}
	return nil
}

func (h *Hosts) Add(id string, handler router.IRouterHandler, service string, method string, append []router.AppendRule) error {
	if len(service) == 0 {
		return h.add(id, handler, router.All, method, append)
	}
	err := h.add(id, handler, service, method, append)
	if err != nil {
		return err
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

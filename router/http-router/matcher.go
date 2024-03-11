package http_router

import (
	"sort"
	"strconv"
	"strings"

	"github.com/eolinker/apinto/checker"
	"github.com/eolinker/apinto/router"
	http_service "github.com/eolinker/eosc/eocontext/http-context"
	"github.com/eolinker/eosc/log"
)

type readerHandler func(port int, request http_service.IRequestReader) (string, bool)

func newPortMatcher(children map[string]router.IMatcher) router.IMatcher {
	return &SimpleMatcher{
		children: children,
		name:     "port",
		read: func(port int, request http_service.IRequestReader) (string, bool) {
			return strconv.Itoa(port), true
		},
	}
}
func newMethodMatcher(children map[string]router.IMatcher, handler router.IRouterHandler) router.IMatcher {
	return &SimpleMatcher{
		children: children,
		name:     "method",
		read: func(port int, request http_service.IRequestReader) (string, bool) {
			return request.Method(), true
		},
	}
}
func newHostMatcher(children map[string]router.IMatcher) router.IMatcher {
	return &SimpleMatcher{
		children: children,
		name:     "host",
		read: func(port int, request http_service.IRequestReader) (string, bool) {
			orgHost := request.URI().Host()
			if i := strings.Index(orgHost, ":"); i > 0 {
				return orgHost[:i], true
			}
			return orgHost, true
		},
	}
}

func newProtocolMatcher(children map[string]router.IMatcher) router.IMatcher {
	return &SimpleMatcher{
		children: children,
		name:     "protocol",
		read: func(port int, request http_service.IRequestReader) (string, bool) {
			return strings.ToLower(request.URI().Scheme()), true
		},
	}
}

type SimpleMatcher struct {
	children map[string]router.IMatcher
	read     readerHandler
	name     string
}

func (s *SimpleMatcher) Match(port int, req interface{}) (router.IRouterHandler, bool) {
	request, ok := req.(http_service.IRequestReader)
	if !ok {
		return nil, false
	}
	log.Debug("SimpleMatcher:", s.name)
	if s == nil || s.children == nil || len(s.children) == 0 {
		return nil, false
	}
	value, _ := s.read(port, request)
	log.Debug("SimpleMatcher:", s.name, "-", value)

	next, has := s.children[value]
	if has {
		handler, ok := next.Match(port, request)
		if ok {
			return handler, true
		}
	}
	next, has = s.children[router.All]
	if has {
		handler, ok := next.Match(port, request)
		if ok {
			return handler, true
		}
	}

	return nil, false

}

type CheckMatcher struct {
	equals   map[string]router.IMatcher //存放使用全等匹配的指标节点
	read     readerHandler
	checkers []*CheckerHandler //按优先顺序存放除全等匹配外的checker，顺序与nodes对应
	all      router.IMatcher
	name     string
}

func NewPathMatcher(equals map[string]router.IMatcher, checkers []*CheckerHandler, all router.IMatcher) *CheckMatcher {
	read := func(port int, request http_service.IRequestReader) (string, bool) {
		return request.URI().Path(), true
	}
	sort.Sort(CheckerSort(checkers))

	return &CheckMatcher{
		name:     "path",
		equals:   equals,
		checkers: checkers,
		read:     read,
		all:      all,
	}
}

func (c *CheckMatcher) Match(port int, req interface{}) (router.IRouterHandler, bool) {
	request, ok := req.(http_service.IRequestReader)
	if !ok {
		return nil, false
	}
	value, hasvalue := c.read(port, request)
	log.Debug("CheckMatcher::Match", "(", len(c.checkers), ")", c.name, "=", value)

	next, has := c.equals[value]
	if has {
		handler, ok := next.Match(port, request)
		if ok {
			return handler, true
		}
	}

	for _, ck := range c.checkers {
		pass := ck.checker.Check(value, hasvalue)
		log.Debug("CheckMatcher::check,", c.name, "=", ck.checker.Key(), pass)

		if pass {
			handler, ok := ck.next.Match(port, request)
			if ok {
				return handler, true
			}
		}
	}
	if c.all != nil {
		return c.all.Match(port, request)
	}
	return nil, false
}

type EmptyMatcher struct {
	handler router.IRouterHandler
	has     bool
}

func (e *EmptyMatcher) Match(port int, request interface{}) (router.IRouterHandler, bool) {
	return e.handler, e.has
}

type AppendMatcher struct {
	handler  router.IRouterHandler
	checkers router.MatcherChecker
}
type AppendMatchers []*AppendMatcher

func (as AppendMatchers) Match(port int, request interface{}) (router.IRouterHandler, bool) {
	for _, m := range as {
		if h, ok := m.Match(port, request); ok {
			return h, true
		}
	}
	return nil, false
}

func (as AppendMatchers) Len() int {
	return len(as)
}

func (as AppendMatchers) Less(i, j int) bool {
	return as[i].checkers.Weight() < as[j].checkers.Weight()
}

func (as AppendMatchers) Swap(i, j int) {
	as[i], as[j] = as[j], as[i]
}

func (a *AppendMatcher) Match(port int, req interface{}) (router.IRouterHandler, bool) {
	request, ok := req.(http_service.IRequestReader)
	if !ok {
		return nil, false
	}
	log.Debug("AppendMatcher")
	if a.checkers.MatchCheck(request) {
		return a.handler, true
	}
	return nil, false
}

type CheckerHandler struct {
	checker checker.Checker
	next    router.IMatcher
}
type CheckerSort []*CheckerHandler

func (cs CheckerSort) Len() int {
	return len(cs)
}

func (cs CheckerSort) Less(i, j int) bool {
	ci, cj := cs[i], cs[j]
	//按匹配规则优先级排序
	if ci.checker.CheckType() != cj.checker.CheckType() {
		return ci.checker.CheckType() < cj.checker.CheckType()
	}

	//按长度排序, 优先级 长>短
	vl := len(ci.checker.Value()) - len(cj.checker.Value())
	if vl != 0 {
		return vl > 0
	}
	return ci.checker.Value() < cj.checker.Value()
}

func (cs CheckerSort) Swap(i, j int) {
	cs[i], cs[j] = cs[j], cs[i]
}

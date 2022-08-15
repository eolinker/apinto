package router

import (
	"github.com/eolinker/apinto/checker"
	http_service "github.com/eolinker/eosc/eocontext/http-context"
	"sort"
	"strconv"
	"strings"
)

const All = "*"

type readerHandler func(port int, request http_service.IRequestReader) (string, bool)

func newPortMatcher(children map[string]IMatcher) IMatcher {
	return &SimpleMatcher{
		children: children,
		read: func(port int, request http_service.IRequestReader) (string, bool) {
			return strconv.Itoa(port), true
		},
	}
}
func newMethodMatcher(children map[string]IMatcher, handler IRouterHandler) IMatcher {
	return &SimpleMatcher{
		children: children,
		read: func(port int, request http_service.IRequestReader) (string, bool) {
			return request.Method(), true
		},
	}
}
func newHostMatcher(children map[string]IMatcher) IMatcher {
	return &SimpleMatcher{
		children: children,
		read: func(port int, request http_service.IRequestReader) (string, bool) {
			orgHost := request.URI().Host()
			if i := strings.Index(orgHost, ":"); i > 0 {
				return orgHost[:i], true
			}
			return orgHost, true
		},
	}
}

type SimpleMatcher struct {
	children map[string]IMatcher
	read     readerHandler
}

func (s *SimpleMatcher) Match(port int, request http_service.IRequestReader) (IRouterHandler, bool) {
	if s == nil || s.children == nil || len(s.children) == 0 {
		return nil, false
	}
	value, _ := s.read(port, request)

	next, has := s.children[value]
	if has {
		handler, ok := next.Match(port, request)
		if ok {
			return handler, true
		}
	}
	next, has = s.children[All]
	if has {
		handler, ok := next.Match(port, request)
		if ok {
			return handler, true
		}
	}

	return nil, false

}

type CheckMatcher struct {
	equals   SimpleMatcher //存放使用全等匹配的指标节点
	read     readerHandler
	checkers []*CheckerHandler //按优先顺序存放除全等匹配外的checker，顺序与nodes对应

}

func NewPathMatcher(equals map[string]IMatcher, checkers []*CheckerHandler) *CheckMatcher {
	read := func(port int, request http_service.IRequestReader) (string, bool) {
		return request.URI().Path(), true
	}
	sort.Sort(CheckerSort(checkers))

	return &CheckMatcher{equals: SimpleMatcher{
		children: equals,
		read:     read,
	}, checkers: checkers,
		read: read,
	}
}

func (c *CheckMatcher) Match(port int, request http_service.IRequestReader) (IRouterHandler, bool) {

	handler, ok := c.equals.Match(port, request)
	if ok {
		return handler, true
	}
	value, has := c.read(port, request)
	for _, ck := range c.checkers {
		pass := ck.checker.Check(value, has)
		if pass {
			handler, ok = ck.next.Match(port, request)
			if ok {
				return handler, true
			}
		}
	}
	return nil, false
}

type EmptyMatcher struct {
	handler IRouterHandler
	has     bool
}

func (e *EmptyMatcher) Match(port int, request http_service.IRequestReader) (IRouterHandler, bool) {
	return e.handler, e.has
}

type AppendMatcher struct {
	handler  IRouterHandler
	checkers MatcherChecker
}
type AppendMatchers []*AppendMatcher

func (as AppendMatchers) Match(port int, request http_service.IRequestReader) (IRouterHandler, bool) {
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
	return as[i].checkers.weight() < as[j].checkers.weight()
}

func (as AppendMatchers) Swap(i, j int) {
	as[i], as[j] = as[j], as[i]
}

func (a *AppendMatcher) Match(port int, request http_service.IRequestReader) (IRouterHandler, bool) {
	if a.checkers.MatchCheck(request) {
		return a.handler, true
	}
	return nil, false
}

type CheckerHandler struct {
	checker checker.Checker
	next    IMatcher
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

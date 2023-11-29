package http_router

import (
	"sort"
	"strings"

	http_service "github.com/eolinker/eosc/eocontext/http-context"

	"github.com/eolinker/apinto/checker"
	"github.com/eolinker/apinto/router"
)

type RuleType = string

const (
	HttpHeader RuleType = "header"
	HttpQuery  RuleType = "query"
	HttpCookie RuleType = "cookie"
)

func Parse(rules []router.AppendRule) router.MatcherChecker {
	if len(rules) == 0 {
		return &router.EmptyChecker{}
	}
	rls := make(router.RuleCheckers, 0, len(rules))

	for _, r := range rules {
		ck, _ := checker.Parse(r.Pattern)

		switch strings.ToLower(r.Type) {
		case HttpHeader:
			rls = append(rls, &HeaderChecker{
				name:    r.Name,
				Checker: ck,
			})
		case HttpQuery:
			rls = append(rls, &QueryChecker{
				name:    r.Name,
				Checker: ck,
			})
		case HttpCookie:
			rls = append(rls, &CookieChecker{
				name:    r.Name,
				Checker: ck,
			})
		}
	}
	sort.Sort(rls)
	return rls
}

type HeaderChecker struct {
	name string
	checker.Checker
}

func (h *HeaderChecker) Weight() int {
	return int(checker.CheckTypeAll-h.Checker.CheckType()) * len(h.Checker.Value())
}

func (h *HeaderChecker) MatchCheck(req interface{}) bool {
	request, ok := req.(http_service.IRequestReader)
	if !ok {
		return false
	}
	v := request.Header().GetHeader(h.name)
	has := len(v) > 0
	return h.Checker.Check(v, has)
}

type CookieChecker struct {
	name string
	checker.Checker
}

func (c *CookieChecker) Weight() int {
	return int(checker.CheckTypeAll-c.Checker.CheckType()) * len(c.Checker.Value())
}

func (c *CookieChecker) MatchCheck(req interface{}) bool {
	request, ok := req.(http_service.IRequestReader)
	if !ok {
		return false
	}
	v := request.Header().GetCookie(c.name)
	has := len(v) > 0
	return c.Checker.Check(v, has)
}

type QueryChecker struct {
	name string
	checker.Checker
}

func (q *QueryChecker) Weight() int {
	return int(checker.CheckTypeAll-q.Checker.CheckType()) * len(q.Checker.Value())
}

func (q *QueryChecker) MatchCheck(req interface{}) bool {
	request, ok := req.(http_service.IRequestReader)
	if !ok {
		return false
	}
	v := request.URI().GetQuery(q.name)
	has := len(v) > 0
	return q.Checker.Check(v, has)
}

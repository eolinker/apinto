package http_router

import (
	"fmt"
	"github.com/eolinker/apinto/checker"
	http_service "github.com/eolinker/eosc/eocontext/http-context"
	"sort"
	"strings"
)

type RuleType = string

const (
	HttpHeader RuleType = "header"
	HttpQuery  RuleType = "query"
	HttpCookie RuleType = "cookie"
)

type AppendRule struct {
	Type    string
	Name    string
	Pattern string
}

func Parse(rules []AppendRule) MatcherChecker {
	if len(rules) == 0 {

		return &EmptyChecker{}
	}
	rls := make(RuleCheckers, len(rules))

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

func Key(rules []AppendRule) string {
	if len(rules) == 0 {
		return All
	}
	rs := make([]string, 0, len(rules))
	for _, r := range rules {
		rs = append(rs, fmt.Sprintf("%s[%s]=%s", strings.ToLower(r.Type), r.Name, strings.TrimSpace(r.Pattern)))
	}
	return strings.Join(rs, "&")
}

type EmptyChecker struct {
}

func (e *EmptyChecker) weight() int {
	return 0
}

func (e *EmptyChecker) MatchCheck(request http_service.IRequestReader) bool {
	return true
}

type HeaderChecker struct {
	name string
	checker.Checker
}

func (h *HeaderChecker) weight() int {
	return int(checker.CheckTypeAll-h.Checker.CheckType()) * len(h.Checker.Value())
}

func (h *HeaderChecker) MatchCheck(request http_service.IRequestReader) bool {
	v := request.Header().GetHeader(h.name)
	has := len(v) > 0
	return h.Checker.Check(v, has)
}

type CookieChecker struct {
	name string
	checker.Checker
}

func (c *CookieChecker) weight() int {
	return int(checker.CheckTypeAll-c.Checker.CheckType()) * len(c.Checker.Value())
}

func (c *CookieChecker) MatchCheck(request http_service.IRequestReader) bool {
	v := request.Header().GetCookie(c.name)
	has := len(v) > 0
	return c.Checker.Check(v, has)
}

type QueryChecker struct {
	name string
	checker.Checker
}

func (q *QueryChecker) weight() int {
	return int(checker.CheckTypeAll-q.Checker.CheckType()) * len(q.Checker.Value())
}

func (q *QueryChecker) MatchCheck(request http_service.IRequestReader) bool {
	v := request.URI().GetQuery(q.name)
	has := len(v) > 0
	return q.Checker.Check(v, has)
}

type MatcherChecker interface {
	MatchCheck(request http_service.IRequestReader) bool
	weight() int
}
type MatcherCheckerItem interface {
	checker.Checker
	MatcherChecker
}
type RuleCheckers []MatcherCheckerItem

func (rs RuleCheckers) weight() int {
	w := len(rs)
	for _, i := range rs {
		w += i.weight()
	}
	return w
}

func (rs RuleCheckers) MatchCheck(request http_service.IRequestReader) bool {
	for _, mc := range rs {
		if !mc.MatchCheck(request) {
			return false
		}
	}
	return true
}

func (rs RuleCheckers) Len() int {
	return len(rs)
}

func (rs RuleCheckers) Less(i, j int) bool {
	ri, rj := rs[i], rs[j]
	//按匹配规则优先级排序
	if ri.CheckType() != rj.CheckType() {
		return ri.CheckType() < rj.CheckType()
	}

	//按长度排序, 优先级 长>短
	vl := len(ri.Value()) - len(rj.Value())
	if vl != 0 {
		return vl > 0
	}
	return ri.Value() < rj.Value()
}

func (rs RuleCheckers) Swap(i, j int) {
	rs[i], rs[j] = rs[j], rs[i]
}

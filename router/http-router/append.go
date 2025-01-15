package http_router

import (
	"sort"
	"strings"

	"github.com/ohler55/ojg/oj"

	"github.com/ohler55/ojg/jp"

	"github.com/eolinker/eosc/log"

	"github.com/eolinker/apinto/checker"
	"github.com/eolinker/apinto/router"
	http_service "github.com/eolinker/eosc/eocontext/http-context"
)

type RuleType = string

const (
	HttpHeader RuleType = "header"
	HttpQuery  RuleType = "query"
	HttpCookie RuleType = "cookie"
	HttpBody   RuleType = "body"
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
		case HttpBody:
			name := r.Name
			if !strings.HasPrefix(r.Name, "$.") {
				name = "$." + r.Name
			}
			expr, err := jp.ParseString(name)
			if err != nil {
				log.Errorf("json path parse error: %v,key is %s", err, r.Name)
				continue
			}
			rls = append(rls, &BodyChecker{
				name:    r.Name,
				expr:    expr,
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

type BodyChecker struct {
	name string
	expr jp.Expr
	checker.Checker
}

func (b *BodyChecker) MatchCheck(req interface{}) bool {
	request, ok := req.(http_service.IRequestReader)
	if !ok {
		return false
	}
	switch request.Body().ContentType() {
	case "application/json":
		body, err := request.Body().RawBody()
		if err != nil {
			log.Errorf("get body error: %v", err)
			return false
		}
		result, err := oj.Parse(body)
		if err != nil {
			log.Errorf("parse body error: %v,body is %s", err, body)
			return false
		}
		return checker.CheckJson(result, checker.JsonArrayMatchAll, b.expr, b.Checker)
	case "application/x-www-form-urlencoded", "multipart/form-data":
		v := request.Body().GetForm(b.name)
		return b.Check(v, len(v) > 0)
	default:
		log.Errorf("unsupported content type: %s", request.Body().ContentType())
		return false
	}
}

func (b *BodyChecker) Weight() int {
	return int(checker.CheckTypeAll-b.Checker.CheckType()) * 50 * len(b.Checker.Value())
}

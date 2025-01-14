package params_check_v2

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/eolinker/apinto/checker"

	"github.com/ohler55/ojg/jp"

	http_service "github.com/eolinker/eosc/eocontext/http-context"
)

var (
	positionHeader = "header"
	positionQuery  = "query"
	positionBody   = "body"
)

type IParamChecker interface {
	Check(logic string, header http_service.IHeaderReader, query http_service.IQueryReader, body interface{}, fn bodyCheckerFunc) bool
}

type headerChecker struct {
	name    string
	checker checker.Checker
}

func (h *headerChecker) Check(logic string, header http_service.IHeaderReader, query http_service.IQueryReader, body interface{}, fn bodyCheckerFunc) bool {
	v := header.GetHeader(h.name)
	match := h.checker.Check(v, true)
	if !match {
		return false
	}
	return true
}

func newHeaderChecker(name string, matchText string) (*headerChecker, error) {
	c, err := checker.Parse(matchText)
	if err != nil {
		return nil, fmt.Errorf("parse param check text error: %w,text: %s", err, matchText)
	}
	return &headerChecker{name: name, checker: c}, nil
}

type queryChecker struct {
	name    string
	checker checker.Checker
}

func (q *queryChecker) Check(logic string, header http_service.IHeaderReader, query http_service.IQueryReader, body interface{}, fn bodyCheckerFunc) bool {
	v := query.GetQuery(q.name)
	match := q.checker.Check(v, true)
	if !match {
		return false
	}
	return true
}

func newQueryChecker(name string, matchText string) (*queryChecker, error) {
	c, err := checker.Parse(matchText)
	if err != nil {
		return nil, fmt.Errorf("parse param check text error: %w,text: %s", err, matchText)
	}
	return &queryChecker{name: name, checker: c}, nil
}

type bodyChecker struct {
	name      string
	expr      jp.Expr
	checker   checker.Checker
	matchMode string
}

func newBodyChecker(name string, matchText string, matchMode string) (*bodyChecker, error) {
	c, _ := checker.Parse(matchText)
	tmp := name
	if !strings.HasPrefix(tmp, "$.") {
		tmp = "$." + tmp
	}
	expr, err := jp.ParseString(tmp)
	if err != nil {
		return nil, err
	}
	return &bodyChecker{
		name:      name,
		expr:      expr,
		checker:   c,
		matchMode: matchMode,
	}, nil
}

func (b *bodyChecker) Check(logic string, header http_service.IHeaderReader, query http_service.IQueryReader, body interface{}, fn bodyCheckerFunc) bool {
	return fn(body, b) == nil
}

type bodyCheckerFunc func(interface{}, *bodyChecker) error

type paramChecker struct {
	logic    string
	checkers []IParamChecker
	checker  IParamChecker
}

func genParamChecker(param *Param) (IParamChecker, error) {
	if param.Position == "" {
		return nil, nil
	}
	var ck IParamChecker
	var err error
	switch param.Position {
	case positionHeader:
		ck, err = newHeaderChecker(param.Name, param.MatchText)
		if err != nil {
			return nil, err
		}
	case positionQuery:
		ck, err = newQueryChecker(param.Name, param.MatchText)
		if err != nil {
			return nil, err
		}
	case positionBody:
		ck, err = newBodyChecker(param.Name, param.MatchText, param.MatchMode)
		if err != nil {
			return nil, err
		}
	}
	return ck, nil
}

func newParamChecker(param *Param) (*paramChecker, error) {
	ck, err := genParamChecker(param)
	if err != nil {
		return nil, err
	}
	cks := make([]IParamChecker, 0, len(param.Params))
	for _, p := range param.Params {
		ck, err := genParamChecker(p)
		if err != nil {
			return nil, err
		}
		cks = append(cks, ck)
	}

	return &paramChecker{
		logic:    param.Logic,
		checker:  ck,
		checkers: cks,
	}, nil
}

func (c *paramChecker) Check(logic string, header http_service.IHeaderReader, query http_service.IQueryReader, body interface{}, fn bodyCheckerFunc) bool {
	success := false
	for _, ck := range c.checkers {
		if ck == nil {
			continue
		}
		ok := ck.Check(c.logic, header, query, body, fn)
		if !ok {
			if logic == logicAnd {
				return false
			}
			continue
		} else {
			success = true
			if logic == logicOr {
				break
			}
		}
	}
	if !success {
		return false
	}
	if c.checker == nil {
		return true
	}
	return c.checker.Check(c.logic, header, query, body, fn)
}

func formChecker(body interface{}, ck *bodyChecker) error {
	params, ok := body.(url.Values)
	if !ok {
		return fmt.Errorf("error body type")
	}
	v := params.Get(ck.name)
	match := ck.checker.Check(v, true)
	if !match {
		return fmt.Errorf(errParamCheck, "body", ck.name)
	}
	return nil
}

func jsonChecker(body interface{}, ck *bodyChecker) error {

	ok := checker.CheckJson(body, ck.matchMode, ck.expr, ck.checker)
	if !ok {
		return fmt.Errorf(errParamCheck, "body", ck.name, ck.name)
	}

	return nil
}

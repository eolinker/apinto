package matcher

import (
	"strconv"
	"strings"

	"github.com/eolinker/eosc/log"

	http_service "github.com/eolinker/eosc/eocontext/http-context"
	"github.com/ohler55/ojg/jp"
	"github.com/ohler55/ojg/oj"
)

func NewJsonMatcher(params []*MatchParam) IMatcher {
	ps := make([]*jsonMatchParam, 0, len(params))
	for _, p := range params {
		key := p.Key
		if !strings.HasPrefix(p.Key, "$.") {
			key = "$." + p.Key
		}
		expr, err := jp.ParseString(key)
		if err != nil {
			log.Errorf("json path parse error: %v,key is %s", err, key)
			continue
		}
		ps = append(ps, &jsonMatchParam{
			MatchParam: p,
			expr:       expr,
		})
	}
	return &jsonMatcher{params: ps}
}

type jsonMatcher struct {
	params []*jsonMatchParam
}

type jsonMatchParam struct {
	*MatchParam
	expr jp.Expr
}

func (m *jsonMatcher) Match(ctx http_service.IHttpContext) bool {
	if len(m.params) < 1 {
		return true
	}
	body := ctx.Response().GetBody()
	tmp, err := oj.Parse(body)
	if err != nil {
		log.Errorf("parse body error: %v,body is %s", err, body)
		return true
	}
	match := true
	for _, p := range m.params {
		results := p.expr.Get(tmp)
		if len(results) < 1 && p.Kind != "nil" {
			return false
		}
		success := true
		for _, r := range results {
			for _, v := range p.Value {

				switch p.Kind {
				case "int":
					t, ok := r.(int64)
					if !ok {
						return false
					}
					val, _ := strconv.ParseInt(v, 10, 64)
					if t != val {
						success = false
						continue
					}
					success = true
					break

				case "bool":
					t, ok := r.(bool)
					if !ok {
						return false
					}
					val, err := strconv.ParseBool(v)
					if err != nil {
						return false
					}
					if t != val {
						success = false
						continue
					}
					success = true
					break
					//return t == val
				default:
					t, ok := r.(string)
					if !ok {
						return false
					}
					if t != v {
						success = false
						continue
					}
					success = true
					break
				}
			}
		}
		if !success {
			return false
		}

	}
	return match
}

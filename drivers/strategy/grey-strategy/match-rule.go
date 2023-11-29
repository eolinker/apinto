package grey_strategy

import (
	"github.com/eolinker/eosc/eocontext"
	http_service "github.com/eolinker/eosc/eocontext/http-context"

	"github.com/eolinker/apinto/checker"
	"github.com/eolinker/apinto/strategy"
)

type ruleGreyMatch struct {
	ruleFilter strategy.IFilter
}

// Match 高级匹配
func (r *ruleGreyMatch) Match(ctx eocontext.EoContext) bool {
	return r.ruleFilter.Check(ctx)
}

type matchingHandler struct {
	Type    string
	name    string
	value   string
	checker checker.Checker
}

type matchingHandlerFilters []*matchingHandler

func (m matchingHandlerFilters) Check(ctx eocontext.EoContext) bool {
	for _, handler := range m {
		if !handler.Check(ctx) {
			return false
		}
	}
	return true
}

func (m *matchingHandler) Check(ctx eocontext.EoContext) bool {
	httpCtx, err := http_service.Assert(ctx)
	if err != nil {
		return false
	}

	value := ""
	request := httpCtx.Request()
	switch m.Type {
	case "header":
		value = request.Header().GetHeader(m.name)
	case "query":
		value = request.URI().GetQuery(m.name)
	case "cookie":
		value = request.Header().GetCookie(m.name)
	default:
		return false
	}

	return m.checker.Check(value, true)
}

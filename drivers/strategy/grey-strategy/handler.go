package grey_strategy

import (
	"github.com/eolinker/apinto/checker"
	"github.com/eolinker/apinto/strategy"
	"github.com/eolinker/eosc/eocontext"
	http_service "github.com/eolinker/eosc/eocontext/http-context"
	"sync"
)

type GreyHandler struct {
	name     string
	filter   strategy.IFilter
	priority int
	stop     bool
	rule     *ruleHandler
}

type greyMatch interface {
	Match(ctx eocontext.EoContext) bool
}

type ruleHandler struct {
	selectNodeLock *sync.Mutex
	index          int
	keepSession    bool

	distribution string
	greyMatch    greyMatch
}

type ruleGreyFlow struct {
	flowRobin *Robin
}

type ruleGreyMatch struct {
	ruleFilter strategy.IFilter
}

func (r *ruleGreyFlow) Match(ctx eocontext.EoContext) bool {
	flow := r.flowRobin.Select()
	return flow.GetId() == 1
}

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

type flowHandler struct {
	id     int //1为灰度流量 2为正常流量
	weight int
}

func (f *flowHandler) GetId() uint32 {
	return uint32(f.id)
}

func (f *flowHandler) GetWeight() int {
	return f.weight
}

func (g *GreyHandler) IsGrey(ctx eocontext.EoContext) bool {
	return g.rule.greyMatch.Match(ctx)
	//cookieKey := fmt.Sprintf(cookieName, g.name)
	//
	//if g.rule.keepSession {
	//	cookie := httpCtx.Request().Header().GetCookie(cookieKey)
	//	if cookie == grey {
	//		return true, nil
	//	} else if cookie == normal {
	//		return false, nil
	//	}
	//}

	//if g.rule.greyMatch.Match(ctx) { //灰度
	//	httpCtx.Response().Headers().Add("Set-Cookie", fmt.Sprintf("%s=%v", cookieKey, grey))
	//	return true, nil
	//} else {
	//	httpCtx.Response().Headers().Add("Set-Cookie", fmt.Sprintf("%s=%v", cookieKey, normal))
	//	return false, nil
	//}
}

func NewGreyHandler(conf *Config) (*GreyHandler, error) {
	filter, err := strategy.ParseFilter(conf.Filters)
	if err != nil {
		return nil, err
	}

	rule := &ruleHandler{
		selectNodeLock: &sync.Mutex{},
		keepSession:    conf.Rule.KeepSession,

		distribution: conf.Rule.Distribution,
	}

	if conf.Rule.Distribution == percent {
		greyFlow := &flowHandler{
			id:     1,
			weight: conf.Rule.Percent,
		}
		normalFlow := &flowHandler{
			id:     2,
			weight: 10000 - greyFlow.weight,
		}
		//总权重10000
		rule.greyMatch = &ruleGreyFlow{flowRobin: NewRobin(greyFlow, normalFlow)}
	} else {
		ruleFilter := make(matchingHandlerFilters, 0)
		for _, matching := range conf.Rule.Matching {

			check, err := checker.Parse(matching.Value)
			if err != nil {
				return nil, err
			}

			matchingHandlerVal := &matchingHandler{
				Type:    matching.Type,
				name:    matching.Name,
				value:   matching.Value,
				checker: check,
			}

			ruleFilter = append(ruleFilter, matchingHandlerVal)
		}

		rule.greyMatch = &ruleGreyMatch{ruleFilter: ruleFilter}
	}

	return &GreyHandler{
		name:     conf.Name,
		filter:   filter,
		priority: conf.Priority,
		stop:     conf.Stop,
		rule:     rule,
	}, nil
}
